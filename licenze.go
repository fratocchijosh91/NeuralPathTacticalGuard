package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	appConfigPath        = "config.json"
	defaultTrialDays     = 7
	licenseVersionCode   = byte(1)
	licenseProductPrefix = "NP"
	licenseTierProName   = "PRO"
)

var licenseSecrets = []string{
	"INSERISCI_IL_TUO_SEGRETO_HMAC_PRIMA_DI_COMPILARE_IN_LOCALE",
}
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

	key, _ := ReadLicenseKey(c)
	key = NormalizeLicenseKey(key)

	if key != "" {
		info, err := ValidateCommercialLicense(key)
		if err == nil && info.Valid {
			return LicenseStatus{
				Mode:        LicenseModePro,
				IsPro:       true,
				IsTrial:     false,
				Valid:       true,
				LicenseKey:  key,
				LicenseID:   info.LicenseID,
				Message:     "Licenza PRO valida",
				ActivatedAt: c.LicenseActivatedAt,
			}
		}

		trial := trialStatus(c, now)
		if trial.Valid {
			trial.Mode = LicenseModeTrial
			trial.LicenseKey = key
			trial.Message = "Chiave non valida, trial ancora attivo"
			return trial
		}

		return LicenseStatus{
			Mode:          LicenseModeInvalid,
			IsPro:         false,
			IsTrial:       false,
			Valid:         false,
			LicenseKey:    key,
			TrialStarted:  c.FirstRunAt,
			TrialFinished: c.FirstRunAt.Add(time.Duration(getTrialDays(c)) * 24 * time.Hour),
			Message:       "Chiave non valida e trial scaduto",
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

	return NormalizeLicenseKey(string(data)), nil
}

func WriteLicenseKey(c *AppConfig, key string) error {
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

	return os.WriteFile(path, []byte(NormalizeLicenseKey(key)), 0644)
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
	info, err := ValidateCommercialLicense(key)
	if err != nil {
		return LicenseStatus{}, err
	}
	if !info.Valid {
		return LicenseStatus{}, errors.New("licenza non valida")
	}

	if err := WriteLicenseKey(c, key); err != nil {
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

func ValidateCommercialLicense(key string) (CommercialLicenseInfo, error) {
	key = NormalizeLicenseKey(key)

	parts := strings.Split(key, "-")
	if len(parts) < 4 {
		return CommercialLicenseInfo{}, errors.New("formato chiave non valido")
	}

	prefix := strings.ToUpper(strings.TrimSpace(parts[0]))
	tier := strings.ToUpper(strings.TrimSpace(parts[1]))
	encoded := strings.Join(parts[2:], "")

	if prefix != licenseProductPrefix {
		return CommercialLicenseInfo{}, fmt.Errorf("prefisso non valido: %s", prefix)
	}
	if tier != licenseTierProName {
		return CommercialLicenseInfo{}, fmt.Errorf("tier non valido: %s", tier)
	}

	raw, err := licenseCrockfordDecode(encoded)
	if err != nil {
		return CommercialLicenseInfo{}, fmt.Errorf("decodifica chiave fallita: %w", err)
	}

	if len(raw) != 12 {
		return CommercialLicenseInfo{}, fmt.Errorf("lunghezza raw non valida: %d", len(raw))
	}

	payload := raw[:7]
	signature := raw[7:]

	if payload[0] != licenseVersionCode {
		return CommercialLicenseInfo{}, errors.New("versione licenza non supportata")
	}

	expectedTierCode, err := licenseTierToCode(tier)
	if err != nil {
		return CommercialLicenseInfo{}, err
	}
	if payload[1] != expectedTierCode {
		return CommercialLicenseInfo{}, errors.New("tier code non valido")
	}

	licenseID := strings.ToUpper(hex.EncodeToString(payload[2:7]))
	if _, revoked := revokedLicenseIDs[licenseID]; revoked {
		return CommercialLicenseInfo{}, errors.New("licenza revocata")
	}

	for _, secret := range licenseSecrets {
		expectedSig := licenseSignPayload(secret, prefix, tier, payload)[:5]
		if hmac.Equal(signature, expectedSig) {
			return CommercialLicenseInfo{
				Valid:     true,
				Prefix:    prefix,
				Tier:      tier,
				LicenseID: licenseID,
			}, nil
		}
	}

	return CommercialLicenseInfo{}, errors.New("firma licenza non valida")
}

func NormalizeLicenseKey(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

func licenseSignPayload(secret, prefix, tier string, payload []byte) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(prefix))
	mac.Write([]byte("|"))
	mac.Write([]byte(tier))
	mac.Write([]byte("|"))
	mac.Write([]byte(hex.EncodeToString(payload)))
	return mac.Sum(nil)
}

func licenseTierToCode(tier string) (byte, error) {
	switch strings.ToUpper(strings.TrimSpace(tier)) {
	case "PRO":
		return 1, nil
	default:
		return 0, fmt.Errorf("tier non supportato: %s", tier)
	}
}

const licenseCrockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

func licenseCrockfordDecode(s string) ([]byte, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "")

	if s == "" {
		return nil, errors.New("stringa vuota")
	}

	lookup := make(map[rune]byte, len(licenseCrockfordAlphabet))
	for i, r := range licenseCrockfordAlphabet {
		lookup[r] = byte(i)
	}

	lookup['O'] = lookup['0']
	lookup['I'] = lookup['1']
	lookup['L'] = lookup['1']

	var out []byte
	var buffer uint32
	var bitsLeft uint8

	for _, r := range s {
		val, ok := lookup[r]
		if !ok {
			return nil, fmt.Errorf("carattere non valido: %q", r)
		}

		buffer = (buffer << 5) | uint32(val)
		bitsLeft += 5

		for bitsLeft >= 8 {
			b := byte(buffer >> (bitsLeft - 8))
			out = append(out, b)
			bitsLeft -= 8

			if bitsLeft > 0 {
				buffer &= (1 << bitsLeft) - 1
			} else {
				buffer = 0
			}
		}
	}

	if len(out) < 12 {
		return nil, fmt.Errorf("dati insufficienti: %d byte", len(out))
	}
	if len(out) > 12 {
		out = out[:12]
	}

	return out, nil
}

/*
   Compatibilità con main.go esistente
*/

func ResolveAccess(args ...interface{}) AccessMode {
	c := pickConfig(args...)
	status := ResolveLicenseStatus(c)

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
	return ResolveLicenseStatus(c).IsPro
}

func ShouldShowUpgradePopup(args ...interface{}) bool {
	c := pickConfig(args...)
	status := ResolveLicenseStatus(c)
	return status.Mode == LicenseModeExpired || status.Mode == LicenseModeInvalid
}

func TrialDaysRemaining(args ...interface{}) int {
	c := pickConfig(args...)
	status := ResolveLicenseStatus(c)
	if status.IsTrial {
		return status.DaysLeft
	}
	return 0
}

func pickConfig(args ...interface{}) *AppConfig {
	if len(args) == 0 {
		return &cfg
	}

	switch v := args[0].(type) {
	case *AppConfig:
		return v
	case AppConfig:
		tmp := v
		return &tmp
	default:
		return &cfg
	}
}
