package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDetectedDevicesFile_Wrapped(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "d.json")
	raw := `{"devices":[{"id":"1","type":"IPHONE","online":true,"last_seen":"2026-04-18T10:00:00Z"}]}`
	if err := os.WriteFile(p, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	got := readDetectedDevicesFile(p)
	if len(got) != 1 || got[0].Type != "IPHONE" || !got[0].Online {
		t.Fatalf("got %#v", got)
	}
}

func TestReadDetectedDevicesFile_Array(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "d.json")
	raw := `[{"id":"a","type":"ANDROID","online":false,"last_seen":"2026-04-18T11:00:00Z"}]`
	if err := os.WriteFile(p, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	got := readDetectedDevicesFile(p)
	if len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("got %#v", got)
	}
}

func TestReadDetectedDevicesFile_Missing(t *testing.T) {
	got := readDetectedDevicesFile(filepath.Join(t.TempDir(), "none.json"))
	if got != nil {
		t.Fatalf("expected nil slice, got %#v", got)
	}
}
