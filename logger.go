package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile *os.File
	logMu   sync.Mutex
)

func initLogger() {
	currentCfg := GetConfig()
	fileName := fmt.Sprintf("neuralpath_log_%s.txt", time.Now().Format("2006-01-02"))
	fullPath := filepath.Join(currentCfg.LogsDir, fileName)

	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Errore creazione log:", err)
		return
	}

	logFile = f
	logEvent("=== AVVIO APPLICAZIONE ===")
}

func logEvent(event string) {
	logMu.Lock()
	defer logMu.Unlock()

	if logFile == nil {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s] %s\n", timestamp, event)
	_, _ = logFile.WriteString(line)
}

func closeLogger() {
	logMu.Lock()
	defer logMu.Unlock()

	if logFile != nil {
		timestamp := time.Now().Format("15:04:05")
		_, _ = logFile.WriteString(fmt.Sprintf("[%s] === CHIUSURA APPLICAZIONE ===\n", timestamp))
		_ = logFile.Close()
		logFile = nil
	}
}
