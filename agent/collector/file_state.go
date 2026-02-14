package collector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type FileState struct {
	Inode  uint64 `json:"inode"`
	Offset int64  `json:"offset"`
}

var (
	stateFilePath = "/var/lib/opslens/state.json"
	stateMu       sync.Mutex
	fileStates    = make(map[string]FileState)
)

func loadState() {
	stateMu.Lock()
	defer stateMu.Unlock()

	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &fileStates)
}

func saveState() {
	stateMu.Lock()
	defer stateMu.Unlock()

	_ = os.MkdirAll(filepath.Dir(stateFilePath), 0755)
	data, _ := json.MarshalIndent(fileStates, "", "  ")

	tmpPath := stateFilePath + ".tmp"

	// Write to temp file first

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return
	}

	// Atomic rename
	_ = os.Rename(tmpPath, stateFilePath)
}
