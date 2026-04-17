package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultAddr         = ":8080"
	defaultProduct      = "neuralpath-tactical-guard"
	defaultTier         = "PRO"
	defaultPrefix       = "NP"
	defaultTokenTTL     = 24 * time.Hour * 30
	maxRequestBodyBytes = 1 << 20
	licenseTokenVersion = 1
)

const defaultRateLimitPerMin = 10

type serverConfig struct {
	addr                string
	product             string
	tier                string
	prefix              string
	allowAnyKey         bool
	tokenTTL            time.Duration
	rateLimitPerMin     int
	allowedKeysPath     string
	adminAPIKey         string
	stripeWebhookSecret string
	allowedKeys         map[string]struct{}
	allowedKeysMu       sync.RWMutex
	privateKey          ed25519.PrivateKey
	publicKey           ed25519.PublicKey
}

type activationRequest struct {
	LicenseKey string `json:"license_key"`
	MachineID  string `json:"machine_id"`
	Product    string `json:"product"`
	Version    string `json:"version"`
}

type activationResponse struct {
	Token   string `json:"token"`
	Message string `json:"message,omitempty"`
}

type signedLicensePayload struct {
	Version   int    `json:"v"`
	Prefix    string `json:"prefix"`
	Tier      string `json:"tier"`
	Product   string `json:"product"`
	LicenseID string `json:"license_id"`
	MachineID string `json:"machine_id"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func main() {
	cfg, err := loadServerConfigFromEnv()
	if err != nil {
		log.Fatalf("configurazione non valida: %v", err)
	}

	activateRL := newIPRateLimiter(cfg.rateLimitPerMin, 1*time.Minute)

	log.Printf("license-server avviato su %s", cfg.addr)
	log.Printf("product=%s tier=%s prefix=%s ttl=%s allowAnyKey=%t allowedKeys=%d rateLimit=%d/min webhookEnabled=%t adminEnabled=%t",
		cfg.product, cfg.tier, cfg.prefix, cfg.tokenTTL.String(), cfg.allowAnyKey, len(cfg.allowedKeys), cfg.rateLimitPerMin, cfg.stripeWebhookSecret != "", cfg.adminAPIKey != "")
	log.Printf("NP_LICENSE_PUBLIC_KEY_B64=%s", base64.StdEncoding.EncodeToString(cfg.publicKey))

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/v1/public-key", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"product":          cfg.product,
			"public_key_b64":   base64.StdEncoding.EncodeToString(cfg.publicKey),
			"signature_scheme": "ed25519",
		})
	})
	mux.HandleFunc("/v1/licenses/activate",
		auditLogMiddleware(rateLimitMiddleware(activateRL, cfg.handleActivate)),
	)
	mux.HandleFunc("/v1/webhooks/stripe", auditLogMiddleware(cfg.handleStripeWebhook))
	mux.HandleFunc("/v1/admin/licenses/create",
		auditLogMiddleware(requireAPIKeyMiddleware(cfg.adminAPIKey, cfg.handleAdminCreateLicense)),
	)

	srv := &http.Server{
		Addr:              cfg.addr,
		Handler:           securityHeadersMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("errore server: %v", err)
	}
}

func (cfg *serverConfig) handleActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, activationResponse{Message: "metodo non supportato"})
		return
	}

	body := http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer func() {
		_ = body.Close()
	}()

	var req activationRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, activationResponse{Message: "payload JSON non valido"})
		return
	}

	req.LicenseKey = normalizeLicenseKey(req.LicenseKey)
	req.MachineID = strings.TrimSpace(req.MachineID)
	req.Product = strings.TrimSpace(req.Product)

	if req.LicenseKey == "" {
		writeJSON(w, http.StatusBadRequest, activationResponse{Message: "license_key obbligatoria"})
		return
	}
	if req.MachineID == "" {
		writeJSON(w, http.StatusBadRequest, activationResponse{Message: "machine_id obbligatorio"})
		return
	}
	if req.Product != "" && req.Product != cfg.product {
		writeJSON(w, http.StatusBadRequest, activationResponse{Message: "product non valido"})
		return
	}
	if !cfg.isAcceptedLicenseKey(req.LicenseKey) {
		writeJSON(w, http.StatusUnauthorized, activationResponse{Message: "chiave licenza non autorizzata"})
		return
	}

	token, err := cfg.buildSignedToken(req)
	if err != nil {
		log.Printf("AUDIT ACTIVATE_FAIL ip=%s key=%s machine=%s err=%v",
			extractIP(r), maskKey(req.LicenseKey), req.MachineID, err)
		writeJSON(w, http.StatusInternalServerError, activationResponse{Message: "errore generazione token"})
		return
	}

	licenseID := deriveLicenseID(req.LicenseKey)
	log.Printf("AUDIT ACTIVATE_OK ip=%s license_id=%s machine=%s product=%s version=%s",
		extractIP(r), licenseID, req.MachineID, req.Product, req.Version)

	writeJSON(w, http.StatusOK, activationResponse{Token: token})
}

func (cfg *serverConfig) buildSignedToken(req activationRequest) (string, error) {
	now := time.Now().UTC()
	licenseID := deriveLicenseID(req.LicenseKey)

	payload := signedLicensePayload{
		Version:   licenseTokenVersion,
		Prefix:    cfg.prefix,
		Tier:      cfg.tier,
		Product:   cfg.product,
		LicenseID: licenseID,
		MachineID: req.MachineID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(cfg.tokenTTL).Unix(),
	}

	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	signature := ed25519.Sign(cfg.privateKey, payloadRaw)
	token := base64.RawURLEncoding.EncodeToString(payloadRaw) + "." + base64.RawURLEncoding.EncodeToString(signature)
	return token, nil
}

func (cfg *serverConfig) isAcceptedLicenseKey(key string) bool {
	parts := strings.Split(key, "-")
	if len(parts) < 2 {
		return false
	}
	if parts[0] != cfg.prefix || parts[1] != cfg.tier {
		return false
	}
	if cfg.allowAnyKey {
		return true
	}
	cfg.allowedKeysMu.RLock()
	defer cfg.allowedKeysMu.RUnlock()

	if len(cfg.allowedKeys) == 0 {
		return false
	}
	_, ok := cfg.allowedKeys[key]
	return ok
}

func loadServerConfigFromEnv() (*serverConfig, error) {
	privateKeyRaw := strings.TrimSpace(os.Getenv("NP_LICENSE_PRIVATE_KEY_B64"))
	if privateKeyRaw == "" {
		return nil, errors.New("NP_LICENSE_PRIVATE_KEY_B64 non impostata")
	}

	privateKey, publicKey, err := parseEd25519PrivateKey(privateKeyRaw)
	if err != nil {
		return nil, err
	}

	cfg := &serverConfig{
		addr:                resolveListenAddr(),
		product:             getEnvOrDefault("NP_LICENSE_PRODUCT", defaultProduct),
		tier:                strings.ToUpper(strings.TrimSpace(getEnvOrDefault("NP_LICENSE_TIER", defaultTier))),
		prefix:              strings.ToUpper(strings.TrimSpace(getEnvOrDefault("NP_LICENSE_PREFIX", defaultPrefix))),
		allowAnyKey:         strings.EqualFold(strings.TrimSpace(os.Getenv("NP_LICENSE_ALLOW_ANY_KEY")), "true"),
		tokenTTL:            parseTTLHours(),
		rateLimitPerMin:     parseRateLimit(),
		allowedKeysPath:     strings.TrimSpace(getEnvOrDefault("NP_LICENSE_KEYS_PATH", "data/allowed-keys.json")),
		adminAPIKey:         strings.TrimSpace(os.Getenv("NP_ADMIN_API_KEY")),
		stripeWebhookSecret: strings.TrimSpace(os.Getenv("NP_STRIPE_WEBHOOK_SECRET")),
		allowedKeys:         parseAllowedLicenseKeys(os.Getenv("NP_LICENSE_KEYS")),
		privateKey:          privateKey,
		publicKey:           publicKey,
	}

	loadedFromFile, err := loadAllowedKeysFromFile(cfg.allowedKeysPath)
	if err != nil {
		log.Printf("warning: impossibile caricare allowed keys da file: %v", err)
	}
	for key := range loadedFromFile {
		cfg.allowedKeys[key] = struct{}{}
	}
	return cfg, nil
}

func resolveListenAddr() string {
	if explicit := strings.TrimSpace(os.Getenv("NP_LICENSE_ADDR")); explicit != "" {
		return explicit
	}

	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return ":" + port
	}

	return defaultAddr
}

func parseRateLimit() int {
	raw := strings.TrimSpace(os.Getenv("NP_LICENSE_RATE_LIMIT_PER_MIN"))
	if raw == "" {
		return defaultRateLimitPerMin
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultRateLimitPerMin
	}
	return n
}

func parseTTLHours() time.Duration {
	raw := strings.TrimSpace(os.Getenv("NP_LICENSE_TOKEN_TTL_HOURS"))
	if raw == "" {
		return defaultTokenTTL
	}
	hours, err := strconv.Atoi(raw)
	if err != nil || hours <= 0 {
		return defaultTokenTTL
	}
	return time.Duration(hours) * time.Hour
}

func parseAllowedLicenseKeys(raw string) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, item := range strings.Split(raw, ",") {
		key := normalizeLicenseKey(item)
		if key == "" {
			continue
		}
		allowed[key] = struct{}{}
	}
	return allowed
}

func parseEd25519PrivateKey(encoded string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	decoded, err := decodeB64Any(encoded)
	if err != nil {
		return nil, nil, fmtError("chiave privata non decodificabile", err)
	}

	switch len(decoded) {
	case ed25519.SeedSize:
		private := ed25519.NewKeyFromSeed(decoded)
		return private, private.Public().(ed25519.PublicKey), nil
	case ed25519.PrivateKeySize:
		private := ed25519.PrivateKey(decoded)
		public := private.Public().(ed25519.PublicKey)
		return private, public, nil
	default:
		return nil, nil, errors.New("chiave privata ed25519 con lunghezza non valida")
	}
}

func decodeB64Any(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("stringa vuota")
	}

	decoders := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	var lastErr error
	for _, enc := range decoders {
		decoded, err := enc.DecodeString(s)
		if err == nil {
			return decoded, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func deriveLicenseID(licenseKey string) string {
	sum := sha256.Sum256([]byte(licenseKey))
	return strings.ToUpper(hex.EncodeToString(sum[:8]))
}

func normalizeLicenseKey(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

func getEnvOrDefault(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:8] + "***"
}

func fmtError(prefix string, err error) error {
	return errors.New(prefix + ": " + err.Error())
}
