package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type AppConfig struct {
	AppTitle           string    `json:"app_title"`
	IPhoneIP           string    `json:"iphone_ip"`
	AndroidIP          string    `json:"android_ip"`
	LagThresholdMs     int       `json:"lag_threshold_ms"`
	RefreshIntervalMs  int       `json:"refresh_interval_ms"`
	ReportsDir         string    `json:"reports_dir"`
	LogsDir            string    `json:"logs_dir"`
	Mode               string    `json:"mode"`
	TrialDays          int       `json:"trial_days"`
	LicenseFile        string    `json:"license_file"`
	FirstRunAt         time.Time `json:"first_run_at"`
	LicenseActivatedAt time.Time `json:"license_activated_at"`
}

var cfg AppConfig

func defaultConfig() AppConfig {
	now := time.Now()

	return AppConfig{
		AppTitle:           "NeuralPath Tactical Guard",
		IPhoneIP:           "172.20.10.1",
		AndroidIP:          "10.145.250.191",
		LagThresholdMs:     100,
		RefreshIntervalMs:  1500,
		ReportsDir:         "reports",
		LogsDir:            "logs",
		Mode:               "real",
		TrialDays:          7,
		LicenseFile:        "license.key",
		FirstRunAt:         now,
		LicenseActivatedAt: time.Time{},
	}
}

func LoadConfig() AppConfig {
	configPath := "config.json"
	c := defaultConfig()
	changed := false

	data, err := os.ReadFile(configPath)
	if err != nil {
		_ = SaveConfig(configPath, c)
		ensureConfigDirs(c)
		cfg = c
		return c
	}

	if err := json.Unmarshal(data, &c); err != nil {
		c = defaultConfig()
		_ = SaveConfig(configPath, c)
		ensureConfigDirs(c)
		cfg = c
		return c
	}

	if c.AppTitle == "" {
		c.AppTitle = "NeuralPath Tactical Guard"
		changed = true
	}
	if c.IPhoneIP == "" {
		c.IPhoneIP = "172.20.10.1"
		changed = true
	}
	if c.AndroidIP == "" {
		c.AndroidIP = "10.145.250.191"
		changed = true
	}
	if c.LagThresholdMs <= 0 {
		c.LagThresholdMs = 100
		changed = true
	}
	if c.RefreshIntervalMs <= 0 {
		c.RefreshIntervalMs = 1500
		changed = true
	}
	if c.ReportsDir == "" {
		c.ReportsDir = "reports"
		changed = true
	}
	if c.LogsDir == "" {
		c.LogsDir = "logs"
		changed = true
	}
	if c.Mode == "" {
		c.Mode = "real"
		changed = true
	}
	if c.TrialDays <= 0 {
		c.TrialDays = 7
		changed = true
	}
	if c.LicenseFile == "" {
		c.LicenseFile = "license.key"
		changed = true
	}
	if c.FirstRunAt.IsZero() {
		c.FirstRunAt = time.Now()
		changed = true
	}

	ensureConfigDirs(c)

	if changed {
		_ = SaveConfig(configPath, c)
	}

	cfg = c
	return c
}

func SaveConfig(path string, c AppConfig) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func ensureConfigDirs(c AppConfig) {
	if c.LogsDir != "" {
		_ = os.MkdirAll(filepath.Clean(c.LogsDir), 0755)
	}
	if c.ReportsDir != "" {
		_ = os.MkdirAll(filepath.Clean(c.ReportsDir), 0755)
	}
	if c.LicenseFile != "" {
		licensePath := filepath.Clean(c.LicenseFile)
		licenseDir := filepath.Dir(licensePath)
		if licenseDir != "." && licenseDir != "" {
			_ = os.MkdirAll(licenseDir, 0755)
		}
	}
}
