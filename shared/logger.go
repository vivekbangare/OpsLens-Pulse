package shared

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func InitLogger(component string) {
	var logPath string

	if runtime.GOOS == "windows" {
		logPath = filepath.Join(
			`C:\ProgramData\OpsLens-Pulse`,
			component+".log",
		)
	} else {
		logPath = filepath.Join(
			"/var/log/opslens-pulse",
			component+".log",
		)
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		log.Fatalf("Failed to create log dir: %v", err)
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, f))
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("📝 Logging initialized → %s\n", logPath)
}
