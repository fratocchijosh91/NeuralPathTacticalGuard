package main

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	appConfigPath           = "config.json"
	defaultTrialDays        = 7
	licenseTierProName      = "PRO"
	licenseProductPrefix    = "NP"
	licenseTokenVersion     = 1
	licenseProductCode      = "neuralpath-tactical-guard"
	licenseActivationPath   = "/v1/licenses/activate"
	licenseRequestTimeout   = 8 * time.Second
	envLicenseServerURL     = "NP_LICENSE_SERVER_URL"
	envLicensePublicKeyBase = "NP_LICENSE_PUBLIC_KEY_B64"
)

var revokedLicenseIDs = map[string]struct{}{
	// "ABCDEF1234": {},
}

type LicenseMode string

const (
	LicenseModeTrial   LicenseMode = "TRIAL"
	LicenseModePro     LicenseMode = "PRO"
	LicenseModeExpired LicenseMode = "EXPIRED"
	LicenseModeInvalid LicenseMode = "INVALID"
)

type LicenseStatus struct {
	Mode          LicenseMode
	IsPro         bool
	IsTrial       bool
	Valid         bool
	ExpiresAt     time.Time
	DaysLeft      int
	LicenseKey    string
	LicenseID     string
	Message       string
	ActivatedAt   time.Time
	TrialStarted  time.Time
	TrialFinished time.Time
}

type CommercialLicenseInfo struct {
	Valid     bool
	Prefix    string
	Tier      string
	LicenseID string
	ExpiresAt time.Time
}

type activationRequest struct {
	LicenseKey string `json:"license_key"`
	MachineID  string `json:"machine_id"`
	Product    string `json:"product"`
	Version    string `json:"version"`
}

type activationResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
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

type AccessMode string

const (
	AccessTrial   AccessMode = "TRIAL"
	AccessPro     AccessMode = "PRO"
	AccessExpired AccessMode = "EXPIRED"
)

func ResolveLicenseStatus(c *AppConfig) LicenseStatus {
	if c == nil {
		return LicenseStatus{
			Mode:    LicenseModeInvalid,
			Valid:   false,
			Message: "config nil",
		}
	}

	now := time.Now()

	if c.FirstRunAt.IsZero() {
		c.FirstRunAt = now
		_ = SaveConfig(appConfigPath, *c)
	}

	token, _ := ReadLicenseKey(c)
	token = normalizeStoredLicenseToken(token)

	if token != "" {
		info, err := validateCommercialLicenseToken(c, token)
		if err == nil && info.Valid {
			return LicenseStatus{
				Mode:        LicenseModePro,
				IsPro:       true,
				IsTrial:     false,
				Valid:       true,
				LicenseKey:  token,
				LicenseID:   info.LicenseID,
				Message:     "Licenza PRO valida",
				ExpiresAt:   info.ExpiresAt,
				ActivatedAt: c.LicenseActivatedAt,
			}
		}

		trial := trialStatus(c, now)
		if trial.Valid {
			trial.Mode = LicenseModeTrial
			trial.LicenseKey = token
			trial.Message = "Token licenza non valido, trial ancora attivo"
			return trial
		}

		return LicenseStatus{
			Mode:          LicenseModeInvalid,
			IsPro:         false,
			IsTrial:       false,
			Valid:         false,
			LicenseKey:    token,
			TrialStarted:  c.FirstRunAt,
			TrialFinished: c.FirstRunAt.Add(time.Duration(getTrialDays(c)) * 24 * time.Hour),
			Message:       "Token licenza non valido e trial scaduto",
		}
	}

	return trialStatus(c, now)
}

func trialStatus(c *AppConfig, now time.Time) LicenseStatus {
	start := c.FirstRunAt
	days := getTrialDays(c)
	end := start.Add(time.Duration(days) * 24 * time.Hour)

	if now.Before(end) {
		daysLeft := int(end.Sub(now).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}

		return LicenseStatus{
			Mode:          LicenseModeTrial,
			IsPro:         false,
			IsTrial:       true,
			Valid:         true,
			ExpiresAt:     end,
			DaysLeft:      daysLeft,
			TrialStarted:  start,
			TrialFinished: end,
			Message:       "Trial attivo",
		}
	}

	return LicenseStatus{
		Mode:          LicenseModeExpired,
		IsPro:         false,
		IsTrial:       false,
		Valid:         false,
		ExpiresAt:     end,
		DaysLeft:      0,
		TrialStarted:  start,
		TrialFinished: end,
		Message:       "Trial scaduto",
	}
}

func getTrialDays(c *AppConfig) int {
	if c == nil || c.TrialDays <= 0 {
		return defaultTrialDays
	}
	return c.TrialDays
}

func ReadLicenseKey(c *AppConfig) (string, error) {
	if c == nil {
		return "", errors.New("config nil")
	}

	path := strings.TrimSpace(c.LicenseFile)
	if path == "" {
		path = "license.key"
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return normalizeStoredLicenseToken(string(data)), nil
}

func WriteLicenseKey(c *AppConfig, token string) error {
	if c == nil {
		return errors.New("config nil")
	}

	path := strings.TrimSpace(c.LicenseFile)
	if path == "" {
		path = "license.key"
	}

	path = filepath.Clean(path)
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(path, []byte(normalizeStoredLicenseToken(token)), 0644)
}

func DeleteLicenseKey(c *AppConfig) error {
	if c == nil {
		return errors.New("config nil")
	}

	path := strings.TrimSpace(c.LicenseFile)
	if path == "" {
		path = "license.key"
	}

	err := os.Remove(filepath.Clean(path))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func ActivateLicense(c *AppConfig, inputKey string) (LicenseStatus, error) {
	if c == nil {
		return LicenseStatus{}, errors.New("config nil")
	}

	key := NormalizeLicenseKey(inputKey)
	if key == "" {
		return LicenseStatus{}, errors.New("chiave licenza vuota")
	}

	token, err := requestActivationToken(c, key)
	if err != nil {
		return LicenseStatus{}, err
	}

	info, err := validateCommercialLicenseToken(c, token)
	if err != nil {
		return LicenseStatus{}, fmt.Errorf("token attivazione non valido: %w", err)
	}
	if !info.Valid {
		return LicenseStatus{}, errors.New("token attivazione non valido")
	}

	if err := WriteLicenseKey(c, token); err != nil {
		return LicenseStatus{}, err
	}

	if c.LicenseActivatedAt.IsZero() {
		c.LicenseActivatedAt = time.Now()
		if err := SaveConfig(appConfigPath, *c); err != nil {
			return LicenseStatus{}, err
		}
	}

	return ResolveLicenseStatus(c), nil
}

func ClearLicense(c *AppConfig) error {
	if c == nil {
		return errors.New("config nil")
	}

	if err := DeleteLicenseKey(c); err != nil {
		return err
	}

	c.LicenseActivatedAt = time.Time{}
	return SaveConfig(appConfigPath, *c)
}

func IsProLicenseActive(c *AppConfig) bool {
	return ResolveLicenseStatus(c).IsPro
}

func ValidateCommercialLicense(token string) (CommercialLicenseInfo, error) {
	currentCfg := pickConfig()
	return validateCommercialLicenseToken(&currentCfg, token)
}

func validateCommercialLicenseToken(c *AppConfig, token string) (CommercialLicenseInfo, error) {
	token = normalizeStoredLicenseToken(token)
	if token == "" {
		return CommercialLicenseInfo{}, errors.New("token licenza vuoto")
	}

	publicKey, err := getLicensePublicKey(c)
	if err != nil {
		return CommercialLicenseInfo{}, err
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return CommercialLicenseInfo{}, errors.New("formato token non valido")
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return CommercialLicenseInfo{}, fmt.Errorf("payload token non valido: %w", err)
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return CommercialLicenseInfo{}, fmt.Errorf("firma token non valida: %w", err)
	}

	if !ed25519.Verify(publicKey, payloadRaw, signature) {
		return CommercialLicenseInfo{}, errors.New("firma token non valida")
	}

	var payload signedLicensePayload
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return CommercialLicenseInfo{}, fmt.Errorf("payload token illeggibile: %w", err)
	}

	if payload.Version != licenseTokenVersion {
		return CommercialLicenseInfo{}, errors.New("versione token non supportata")
	}
	if strings.ToUpper(strings.TrimSpace(payload.Prefix)) != licenseProductPrefix {
		return CommercialLicenseInfo{}, errors.New("prefisso token non valido")
	}
	if strings.ToUpper(strings.TrimSpace(payload.Tier)) != licenseTierProName {
		return CommercialLicenseInfo{}, errors.New("tier token non valido")
	}
	if payload.Product != "" && payload.Product != licenseProductCode {
		return CommercialLicenseInfo{}, errors.New("prodotto token non valido")
	}

	licenseID := strings.ToUpper(strings.TrimSpace(payload.LicenseID))
	if licenseID == "" {
		return CommercialLicenseInfo{}, errors.New("license_id mancante")
	}
	if _, revoked := revokedLicenseIDs[licenseID]; revoked {
		return CommercialLicenseInfo{}, errors.New("licenza revocata")
	}

	expiresAt := time.Unix(payload.ExpiresAt, 0).UTC()
	if payload.ExpiresAt <= 0 || time.Now().UTC().After(expiresAt) {
		return CommercialLicenseInfo{}, errors.New("token scaduto")
	}

	machineID := getMachineID()
	if payload.MachineID != "" && payload.MachineID != machineID {
		return CommercialLicenseInfo{}, errors.New("token associato a un altro dispositivo")
	}

	return CommercialLicenseInfo{
		Valid:     true,
		Prefix:    strings.ToUpper(strings.TrimSpace(payload.Prefix)),
		Tier:      strings.ToUpper(strings.TrimSpace(payload.Tier)),
		LicenseID: licenseID,
		ExpiresAt: expiresAt,
	}, nil
}

func NormalizeLicenseKey(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

func normalizeStoredLicenseToken(s string) string {
	return strings.TrimSpace(s)
}

func requestActivationToken(c *AppConfig, key string) (string, error) {
	serverURL := getLicenseServerURL(c)
	if serverURL == "" {
		return "", fmt.Errorf("server licenze non configurato (usa config.%s o env %s)", "license_server_url", envLicenseServerURL)
	}

	reqBody := activationRequest{
		LicenseKey: key,
		MachineID:  getMachineID(),
		Product:    licenseProductCode,
		Version:    version,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), licenseRequestTimeout)
	defer cancel()

	endpoint := strings.TrimRight(serverURL, "/") + licenseActivationPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("richiesta attivazione fallita: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if len(body) > 0 {
			return "", fmt.Errorf("attivazione rifiutata: %s", strings.TrimSpace(string(body)))
		}
		return "", fmt.Errorf("attivazione rifiutata: HTTP %d", resp.StatusCode)
	}

	var activation activationResponse
	if err := json.Unmarshal(body, &activation); err != nil {
		return "", fmt.Errorf("risposta attivazione non valida: %w", err)
	}
	if strings.TrimSpace(activation.Token) == "" {
		if activation.Message != "" {
			return "", errors.New(activation.Message)
		}
		return "", errors.New("token attivazione mancante")
	}

	return normalizeStoredLicenseToken(activation.Token), nil
}

func getLicenseServerURL(c *AppConfig) string {
	if c != nil && strings.TrimSpace(c.LicenseServerURL) != "" {
		return strings.TrimSpace(c.LicenseServerURL)
	}
	return strings.TrimSpace(os.Getenv(envLicenseServerURL))
}

func getLicensePublicKey(c *AppConfig) (ed25519.PublicKey, error) {
	encoded := ""
	if c != nil {
		encoded = strings.TrimSpace(c.LicensePublicKey)
	}
	if encoded == "" {
		encoded = strings.TrimSpace(os.Getenv(envLicensePublicKeyBase))
	}
	if encoded == "" {
		return nil, fmt.Errorf("chiave pubblica licenze non configurata (config.%s o env %s)", "license_public_key", envLicensePublicKeyBase)
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(encoded)
	}
	if err != nil {
		raw, err = base64.RawURLEncoding.DecodeString(encoded)
	}
	if err != nil {
		return nil, fmt.Errorf("chiave pubblica non valida: %w", err)
	}

	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("dimensione chiave pubblica non valida: %d", len(raw))
	}
	return ed25519.PublicKey(raw), nil
}

func getMachineID() string {
	hostName, _ := os.Hostname()
	fingerprint := fmt.Sprintf("%s|%s|%s", strings.ToLower(strings.TrimSpace(hostName)), runtime.GOOS, runtime.GOARCH)
	sum := sha256.Sum256([]byte(fingerprint))
	return strings.ToUpper(hex.EncodeToString(sum[:8]))
}

/*
   Compatibilità con main.go esistente
*/

func ResolveAccess(args ...interface{}) AccessMode {
	c := pickConfig(args...)
	status := ResolveLicenseStatus(&c)

	if status.IsPro {
		return AccessPro
	}
	if status.Mode == LicenseModeTrial {
		return AccessTrial
	}
	return AccessExpired
}

func CanExportReport(args ...interface{}) bool {
	c := pickConfig(args...)
	return ResolveLicenseStatus(&c).IsPro
}

func ShouldShowUpgradePopup(args ...interface{}) bool {
	c := pickConfig(args...)
	status := ResolveLicenseStatus(&c)
	return status.Mode == LicenseModeExpired || status.Mode == LicenseModeInvalid
}

func TrialDaysRemaining(args ...interface{}) int {
	c := pickConfig(args...)
	status := ResolveLicenseStatus(&c)
	if status.IsTrial {
		return status.DaysLeft
	}
	return 0
}

func pickConfig(args ...interface{}) AppConfig {
	if len(args) == 0 {
		return GetConfig()
	}

	switch v := args[0].(type) {
	case *AppConfig:
		if v == nil {
			return GetConfig()
		}
		return *v
	case AppConfig:
		return v
	default:
		return GetConfig()
	}
}
