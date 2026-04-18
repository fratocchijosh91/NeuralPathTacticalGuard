package main

import (
	"encoding/json"
	"net/http"
	"os"
)

// detectedDevice rappresenta un evento di rilevamento hotspot (popolato da file JSON sul server).
type detectedDevice struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Online     bool   `json:"online"`
	StarlinkMs int    `json:"starlink_ms,omitempty"`
	DeviceMs   int    `json:"device_ms,omitempty"`
	LastSeen   string `json:"last_seen"`
}

type detectedDevicesResponse struct {
	Devices []detectedDevice `json:"devices"`
}

func (cfg *serverConfig) handleDetectedDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "metodo non supportato"})
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	devices := readDetectedDevicesFile(cfg.detectedDevicesPath)
	writeJSON(w, http.StatusOK, detectedDevicesResponse{Devices: devices})
}

func readDetectedDevicesFile(path string) []detectedDevice {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var wrapped detectedDevicesResponse
	if err := json.Unmarshal(raw, &wrapped); err == nil && wrapped.Devices != nil {
		return wrapped.Devices
	}
	var arr []detectedDevice
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr
	}
	return nil
}
