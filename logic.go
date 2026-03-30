package main

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

func resetLogicState() {
	lastStatus = false
	lastStatusInitialized = false
}

func updateLogic(net Network, prevAlarms int) State {
	device, online := net.DetectDevice()

	starlinkPing := net.Ping("8.8.8.8")
	devicePing := 999

	if online {
		if device == "IPHONE" {
			devicePing = net.Ping(cfg.IPhoneIP)
		} else if device == "ANDROID" {
			devicePing = net.Ping(cfg.AndroidIP)
		}
	}

	alarms := prevAlarms

	if !lastStatusInitialized {
		lastStatus = online
		lastStatusInitialized = true
	} else if online != lastStatus {
		if !online {
			alarms++
		}
		lastStatus = online
	}

	return State{
		StarlinkPing: starlinkPing,
		DevicePing:   devicePing,
		IsOnline:     online,
		Device:       device,
		AlarmCount:   alarms,
	}
}
