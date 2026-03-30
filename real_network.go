package main

import (
	"net"
	"time"
)

type RealNetwork struct{}

func (r RealNetwork) Ping(host string) int {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", host+":53", 1*time.Second)
	if err != nil {
		return 999
	}
	defer conn.Close()

	return int(time.Since(start).Milliseconds())
}

func (r RealNetwork) DetectDevice() (string, bool) {
	conn, err := net.DialTimeout("tcp", cfg.IPhoneIP+":53", 150*time.Millisecond)
	if err == nil {
		conn.Close()
		return "IPHONE", true
	}

	conn, err = net.DialTimeout("tcp", cfg.AndroidIP+":53", 150*time.Millisecond)
	if err == nil {
		conn.Close()
		return "ANDROID", true
	}

	return "NONE", false
}
