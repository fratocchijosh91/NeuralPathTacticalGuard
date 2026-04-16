package main

import "sync"

type MockNetwork struct {
	mu           sync.RWMutex
	Online       bool
	Device       string
	StarlinkPing int
	DevicePing   int
}

func (m *MockNetwork) Ping(host string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	currentCfg := GetConfig()

	switch host {
	case "8.8.8.8":
		return m.StarlinkPing
	case currentCfg.IPhoneIP, currentCfg.AndroidIP:
		return m.DevicePing
	default:
		return 999
	}
}

func (m *MockNetwork) DetectDevice() (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Device, m.Online
}

func (m *MockNetwork) SetState(online bool, device string, starlinkPing int, devicePing int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Online = online
	m.Device = device
	m.StarlinkPing = starlinkPing
	m.DevicePing = devicePing
}
