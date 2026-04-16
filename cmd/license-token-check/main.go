package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

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
	token := flag.String("token", "", "token licenza da verificare")
	publicKeyB64 := flag.String("public-key-b64", "", "chiave pubblica ed25519 base64")
	expectedProduct := flag.String("expected-product", "neuralpath-tactical-guard", "product atteso")
	expectedPrefix := flag.String("expected-prefix", "NP", "prefix atteso")
	expectedTier := flag.String("expected-tier", "PRO", "tier atteso")
	expectedMachine := flag.String("expected-machine-id", "", "machine id atteso (opzionale)")
	flag.Parse()

	if err := run(*token, *publicKeyB64, *expectedProduct, *expectedPrefix, *expectedTier, *expectedMachine); err != nil {
		fmt.Fprintf(os.Stderr, "token non valido: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("token valido")
}

func run(token, publicKeyB64, expectedProduct, expectedPrefix, expectedTier, expectedMachine string) error {
	token = strings.TrimSpace(token)
	publicKeyB64 = strings.TrimSpace(publicKeyB64)
	if token == "" {
		return errors.New("token vuoto")
	}
	if publicKeyB64 == "" {
		return errors.New("public key vuota")
	}

	publicKeyRaw, err := decodeB64Any(publicKeyB64)
	if err != nil {
		return fmt.Errorf("public key non decodificabile: %w", err)
	}
	if len(publicKeyRaw) != ed25519.PublicKeySize {
		return fmt.Errorf("public key con lunghezza non valida: %d", len(publicKeyRaw))
	}
	publicKey := ed25519.PublicKey(publicKeyRaw)

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return errors.New("formato token non valido")
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("payload token non valido: %w", err)
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("signature token non valida: %w", err)
	}

	if !ed25519.Verify(publicKey, payloadRaw, signature) {
		return errors.New("firma token non valida")
	}

	var payload signedLicensePayload
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return fmt.Errorf("payload non decodificabile: %w", err)
	}

	if strings.TrimSpace(payload.Product) != strings.TrimSpace(expectedProduct) {
		return fmt.Errorf("product non valido: %s", payload.Product)
	}
	if strings.ToUpper(strings.TrimSpace(payload.Prefix)) != strings.ToUpper(strings.TrimSpace(expectedPrefix)) {
		return fmt.Errorf("prefix non valido: %s", payload.Prefix)
	}
	if strings.ToUpper(strings.TrimSpace(payload.Tier)) != strings.ToUpper(strings.TrimSpace(expectedTier)) {
		return fmt.Errorf("tier non valido: %s", payload.Tier)
	}
	if strings.TrimSpace(payload.LicenseID) == "" {
		return errors.New("license_id mancante")
	}

	expiresAt := time.Unix(payload.ExpiresAt, 0).UTC()
	if payload.ExpiresAt <= 0 || time.Now().UTC().After(expiresAt) {
		return errors.New("token scaduto")
	}
	if expectedMachine != "" && payload.MachineID != expectedMachine {
		return fmt.Errorf("machine_id non valido: %s", payload.MachineID)
	}

	return nil
}

func decodeB64Any(s string) ([]byte, error) {
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
