package main

import (
	"context"
	"net"
	"strings"
	"time"
)

const (
	defaultPingTimeout         = 1 * time.Second
	defaultDetectDeviceTimeout = 150 * time.Millisecond
	defaultProbePort           = "53"
)

type RealNetwork struct{}

func (r RealNetwork) Ping(host string) int {
	start := time.Now()

	conn, err := dialTCP(context.Background(), host, defaultPingTimeout)
	if err != nil {
		return 999
	}
	_ = conn.Close()

	return int(time.Since(start).Milliseconds())
}

func (r RealNetwork) DetectDevice() (string, bool) {
	currentCfg := GetConfig()

	conn, err := dialTCP(context.Background(), currentCfg.IPhoneIP, defaultDetectDeviceTimeout)
	if err == nil {
		_ = conn.Close()
		return "IPHONE", true
	}

	conn, err = dialTCP(context.Background(), currentCfg.AndroidIP, defaultDetectDeviceTimeout)
	if err == nil {
		_ = conn.Close()
		return "ANDROID", true
	}

	return "NONE", false
}

func dialTCP(ctx context.Context, host string, timeout time.Duration) (net.Conn, error) {
	target := normalizeTarget(host)
	if target == "" {
		return nil, net.InvalidAddrError("host vuoto")
	}

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var dialer net.Dialer
	return dialer.DialContext(dialCtx, "tcp", target)
}

func normalizeTarget(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}

	if _, _, err := net.SplitHostPort(host); err == nil {
		return host
	}

	return net.JoinHostPort(host, defaultProbePort)
}
