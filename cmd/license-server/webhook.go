package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type stripeEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object stripeCheckoutSession `json:"object"`
	} `json:"data"`
}

type stripeCheckoutSession struct {
	ID              string            `json:"id"`
	PaymentStatus   string            `json:"payment_status"`
	CustomerEmail   string            `json:"customer_email"`
	ClientReference string            `json:"client_reference_id"`
	Metadata        map[string]string `json:"metadata"`
}

type allowedKeysFile struct {
	Keys []string `json:"keys"`
}

func (cfg *serverConfig) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "metodo non supportato"})
		return
	}
	if cfg.stripeWebhookSecret == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "webhook Stripe non configurato"})
		return
	}

	bodyReader := http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer func() { _ = bodyReader.Close() }()

	payload, err := io.ReadAll(bodyReader)
	if err != nil {
		writeJSON(w, 400, map[string]string{"message": "payload non valido"})
		return
	}
	if !cfg.verifyStripeSignature(payload, r.Header.Get("Stripe-Signature")) {
		writeJSON(w, 401, map[string]string{"message": "firma webhook non valida"})
		return
	}

	var event stripeEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		writeJSON(w, 400, map[string]string{"message": "evento JSON non valido"})
		return
	}

	if event.Type != "checkout.session.completed" {
		writeJSON(w, 200, map[string]string{"status": "ignored", "event": event.Type})
		return
	}

	session := event.Data.Object
	if !strings.EqualFold(strings.TrimSpace(session.PaymentStatus), "paid") {
		writeJSON(w, 200, map[string]string{"status": "ignored", "reason": "payment_status_not_paid"})
		return
	}

	licenseKey := normalizeLicenseKey(session.Metadata["license_key"])
	if licenseKey == "" {
		licenseKey = cfg.generateLicenseKeyFromSession(session)
	}

	if !cfg.matchesKeySchema(licenseKey) {
		writeJSON(w, 400, map[string]string{"message": "license_key non valida per prefix/tier"})
		return
	}

	if err := cfg.addAllowedKey(licenseKey); err != nil {
		writeJSON(w, 500, map[string]string{"message": "errore salvataggio licenza"})
		return
	}

	log.Printf("AUDIT STRIPE_WEBHOOK_OK event_id=%s session_id=%s email=%s key=%s",
		event.ID, session.ID, session.CustomerEmail, maskKey(licenseKey))

	writeJSON(w, 200, map[string]string{
		"status":      "ok",
		"license_key": licenseKey,
	})
}

func (cfg *serverConfig) verifyStripeSignature(payload []byte, signatureHeader string) bool {
	ts, sig := parseStripeSignatureHeader(signatureHeader)
	if ts == "" || sig == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(cfg.stripeWebhookSecret))
	_, _ = mac.Write([]byte(ts))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

func parseStripeSignatureHeader(header string) (timestamp string, signature string) {
	parts := strings.Split(header, ",")
	for _, part := range parts {
		piece := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(piece) != 2 {
			continue
		}
		switch piece[0] {
		case "t":
			timestamp = piece[1]
		case "v1":
			signature = piece[1]
		}
	}
	return timestamp, signature
}

func (cfg *serverConfig) generateLicenseKeyFromSession(session stripeCheckoutSession) string {
	base := strings.TrimSpace(session.ClientReference)
	if base == "" {
		base = strings.TrimSpace(session.CustomerEmail)
	}
	if base == "" {
		base = strings.TrimSpace(session.ID)
	}
	if base == "" {
		base = fmt.Sprintf("AUTO-%d", time.Now().Unix())
	}
	return cfg.generateLicenseKeyFromSeed(base)
}

func (cfg *serverConfig) matchesKeySchema(key string) bool {
	parts := strings.Split(key, "-")
	if len(parts) < 3 {
		return false
	}
	return parts[0] == cfg.prefix && parts[1] == cfg.tier
}

func (cfg *serverConfig) addAllowedKey(key string) error {
	cfg.allowedKeysMu.Lock()
	cfg.allowedKeys[key] = struct{}{}
	keys := make([]string, 0, len(cfg.allowedKeys))
	for k := range cfg.allowedKeys {
		keys = append(keys, k)
	}
	cfg.allowedKeysMu.Unlock()

	sort.Strings(keys)
	return saveAllowedKeysToFile(cfg.allowedKeysPath, keys)
}

func loadAllowedKeysFromFile(path string) (map[string]struct{}, error) {
	keys := make(map[string]struct{})
	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return keys, nil
		}
		return nil, err
	}

	var payload allowedKeysFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	for _, k := range payload.Keys {
		nk := normalizeLicenseKey(k)
		if nk != "" {
			keys[nk] = struct{}{}
		}
	}
	return keys, nil
}

func saveAllowedKeysToFile(path string, keys []string) error {
	cleanPath := filepath.Clean(path)
	dir := filepath.Dir(cleanPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	payload := allowedKeysFile{Keys: keys}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cleanPath, data, 0o600)
}
