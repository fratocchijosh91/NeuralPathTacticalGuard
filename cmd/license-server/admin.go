package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type adminCreateLicenseRequest struct {
	Reference  string `json:"reference"`
	Email      string `json:"email"`
	LicenseKey string `json:"license_key"`
}

type adminCreateLicenseResponse struct {
	Status     string `json:"status"`
	LicenseKey string `json:"license_key"`
	Message    string `json:"message,omitempty"`
}

func (cfg *serverConfig) handleAdminCreateLicense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, adminCreateLicenseResponse{Message: "metodo non supportato"})
		return
	}

	body := http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer func() {
		_ = body.Close()
	}()

	var req adminCreateLicenseRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, adminCreateLicenseResponse{Message: "payload JSON non valido"})
		return
	}

	req.Reference = strings.TrimSpace(req.Reference)
	req.Email = strings.TrimSpace(req.Email)
	req.LicenseKey = normalizeLicenseKey(req.LicenseKey)

	licenseKey := req.LicenseKey
	if licenseKey == "" {
		seed := req.Reference
		if seed == "" {
			seed = req.Email
		}
		if seed == "" {
			seed = fmt.Sprintf("admin-%d", time.Now().UnixNano())
		}
		licenseKey = cfg.generateLicenseKeyFromSeed(seed)
	}

	if !cfg.matchesKeySchema(licenseKey) {
		writeJSON(w, http.StatusBadRequest, adminCreateLicenseResponse{
			Message: "license_key non valida per prefix/tier",
		})
		return
	}

	if err := cfg.addAllowedKey(licenseKey); err != nil {
		writeJSON(w, http.StatusInternalServerError, adminCreateLicenseResponse{
			Message: "errore salvataggio licenza",
		})
		return
	}

	log.Printf("AUDIT ADMIN_CREATE_LICENSE ip=%s key=%s ref=%s email=%s",
		extractIP(r), maskKey(licenseKey), req.Reference, req.Email)

	writeJSON(w, http.StatusOK, adminCreateLicenseResponse{
		Status:     "ok",
		LicenseKey: licenseKey,
	})
}

func (cfg *serverConfig) generateLicenseKeyFromSeed(seed string) string {
	seed = strings.TrimSpace(seed)
	if seed == "" {
		seed = fmt.Sprintf("auto-%d", time.Now().UnixNano())
	}
	sum := sha256.Sum256([]byte(seed))
	code := strings.ToUpper(hex.EncodeToString(sum[:6]))
	return fmt.Sprintf("%s-%s-%s", cfg.prefix, cfg.tier, code)
}
