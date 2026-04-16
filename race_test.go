package main

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

type flappingNetwork struct {
	toggles atomic.Int64
}

func (f *flappingNetwork) Ping(host string) int {
	currentCfg := GetConfig()

	switch host {
	case "8.8.8.8":
		return 30
	case currentCfg.IPhoneIP:
		return 12
	case currentCfg.AndroidIP:
		return 18
	default:
		return 999
	}
}

func (f *flappingNetwork) DetectDevice() (string, bool) {
	n := f.toggles.Add(1)
	if n%2 == 0 {
		return "IPHONE", true
	}
	return "NONE", false
}

func TestUpdateLogicConcurrentAccess(t *testing.T) {
	oldCfg := GetConfig()
	defer SetConfig(oldCfg)

	testCfg := defaultConfig()
	SetConfig(testCfg)

	resetLogicState()
	defer resetLogicState()

	netStub := &flappingNetwork{}

	const workers = 8
	const iterations = 500

	var wg sync.WaitGroup
	errCh := make(chan string, workers*iterations)

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			alarms := 0
			for i := 0; i < iterations; i++ {
				state := updateLogic(netStub, alarms)
				alarms = state.AlarmCount

				if state.IsOnline && state.DevicePing == 999 {
					errCh <- "online con ping device non valido"
					return
				}
				if !state.IsOnline && state.Device != "NONE" {
					errCh <- "offline con device inatteso"
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for errMsg := range errCh {
		t.Fatalf("stato inconsistente: %s", errMsg)
	}
}

func TestConfigConcurrentReadWrite(t *testing.T) {
	oldCfg := GetConfig()
	defer SetConfig(oldCfg)

	SetConfig(defaultConfig())

	const writers = 6
	const readers = 12
	const iterations = 400

	var wg sync.WaitGroup
	errCh := make(chan error, writers*iterations)

	for writer := 0; writer < writers; writer++ {
		writerID := writer
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				err := WithConfigWrite(func(c *AppConfig) error {
					c.LagThresholdMs = 80 + writerID + i
					c.RefreshIntervalMs = 300 + writerID + i
					return nil
				})
				if err != nil {
					errCh <- err
					return
				}
			}
		}()
	}

	for reader := 0; reader < readers; reader++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				currentCfg := GetConfig()
				if currentCfg.AppTitle == "" {
					errCh <- errors.New("app title vuoto")
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("errore accesso concorrente config: %v", err)
		}
	}

	finalCfg := GetConfig()
	if finalCfg.LagThresholdMs <= 0 || finalCfg.RefreshIntervalMs <= 0 {
		t.Fatalf("config finale non valida: %+v", finalCfg)
	}
}
