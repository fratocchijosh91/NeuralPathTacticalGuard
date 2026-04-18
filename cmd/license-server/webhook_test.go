package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestParseStripeSignatureHeader_MultipleV1(t *testing.T) {
	header := "t=1234567890,v1=aaa,v1=bbb"
	ts, sigs := parseStripeSignatureHeader(header)
	if ts != "1234567890" {
		t.Fatalf("timestamp: got %q", ts)
	}
	if len(sigs) != 2 || sigs[0] != "aaa" || sigs[1] != "bbb" {
		t.Fatalf("signatures: %#v", sigs)
	}
}

func TestVerifyStripeSignature_OK(t *testing.T) {
	secret := "whsec_test_plaintext_dev_secret"
	payload := []byte(`{"id":"evt_1","type":"checkout.session.completed"}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ts))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	header := "t=" + ts + ",v1=" + sig
	cfg := &serverConfig{
		stripeWebhookSecret:    secret,
		stripeWebhookTolerance: 5 * time.Minute,
	}
	if !cfg.verifyStripeSignature(payload, header) {
		t.Fatal("expected signature valid")
	}
}

func TestVerifyStripeSignature_ReplayRejected(t *testing.T) {
	secret := "whsec_test_plaintext_dev_secret"
	payload := []byte(`{"id":"evt_1"}`)
	ts := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ts))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	header := "t=" + ts + ",v1=" + sig
	cfg := &serverConfig{
		stripeWebhookSecret:    secret,
		stripeWebhookTolerance: 5 * time.Minute,
	}
	if cfg.verifyStripeSignature(payload, header) {
		t.Fatal("expected replay outside tolerance to fail")
	}
}

func TestVerifyStripeSignature_MatchesSecondV1(t *testing.T) {
	secret := "whsec_test_plaintext_dev_secret"
	payload := []byte(`{"x":1}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ts))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	good := strings.ToLower(hex.EncodeToString(mac.Sum(nil)))

	header := "t=" + ts + ",v1=deadbeef,v1=" + good
	cfg := &serverConfig{
		stripeWebhookSecret:    secret,
		stripeWebhookTolerance: 5 * time.Minute,
	}
	if !cfg.verifyStripeSignature(payload, header) {
		t.Fatal("expected second v1 signature to match")
	}
}
