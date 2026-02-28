package collector

import (
	"os"
	"path/filepath"
	"runtime"
)

var BaseDir string
var QueueDir string
var StateDir string
var LogDir string

func init() {

	if runtime.GOOS == "windows" {
		BaseDir = `C:\ProgramData\OpsLens-Pulse`
	} else {

		// Root user → system location
		if os.Geteuid() == 0 {
			BaseDir = "/var/lib/opslens-pulse"
		} else {
			// Non-root → user home
			home, err := os.UserHomeDir()
			if err != nil {
				BaseDir = "./.opslens-pulse"
			} else {
				BaseDir = filepath.Join(home, ".opslens-pulse")
			}
		}
	}

	QueueDir = filepath.Join(BaseDir, "queue")
	StateDir = filepath.Join(BaseDir, "state")
	LogDir = filepath.Join(BaseDir, "logs")

	// Ensure directories exist
	_ = os.MkdirAll(QueueDir, 0700)
	_ = os.MkdirAll(StateDir, 0700)
	_ = os.MkdirAll(LogDir, 0700)
}
