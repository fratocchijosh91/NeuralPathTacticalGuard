package main

import "sync"

type Network interface {
	Ping(host string) int
	DetectDevice() (string, bool)
}

type State struct {
	StarlinkPing int
	DevicePing   int
	IsOnline     bool
	Device       string
	AlarmCount   int
}

var lastStatus bool
var lastStatusInitialized bool
var logicStateMu sync.Mutex

func resetLogicState() {
	logicStateMu.Lock()
	defer logicStateMu.Unlock()

	lastStatus = false
	lastStatusInitialized = false
}

func updateLogic(net Network, prevAlarms int) State {
	device, online := net.DetectDevice()
	currentCfg := GetConfig()

	starlinkPing := net.Ping("8.8.8.8")
	devicePing := 999

	if online {
		if device == "IPHONE" {
			devicePing = net.Ping(currentCfg.IPhoneIP)
		} else if device == "ANDROID" {
			devicePing = net.Ping(currentCfg.AndroidIP)
		}
	}

	alarms := prevAlarms

	logicStateMu.Lock()
	if !lastStatusInitialized {
		lastStatus = online
		lastStatusInitialized = true
	} else if online != lastStatus {
		if !online {
			alarms++
		}
		lastStatus = online
	}
	logicStateMu.Unlock()

	return State{
		StarlinkPing: starlinkPing,
		DevicePing:   devicePing,
		IsOnline:     online,
		Device:       device,
		AlarmCount:   alarms,
	}
}
