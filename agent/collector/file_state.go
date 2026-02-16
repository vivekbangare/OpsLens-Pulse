package collector

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileState struct {
	Inode  uint64 `json:"inode"`
	Offset int64  `json:"offset"`
}

var (
	stateFilePath = filepath.Join(BaseDir, "state.json")
	stateMu       sync.Mutex
	fileStates    = make(map[string]FileState)

	stateDirty      = false
	stateFlushEvery = 30 * time.Second
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

	if !stateDirty {
		return
	}

	_ = os.MkdirAll(filepath.Dir(stateFilePath), 0755)

	data, _ := json.MarshalIndent(fileStates, "", "  ")

	tmp := stateFilePath + ".tmp"

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return
	}

	_ = os.Rename(tmp, stateFilePath)

	stateDirty = false
}

// periodicStateFlush runs background flusher
func periodicStateFlush(ctx context.Context) {
	ticker := time.NewTicker(stateFlushEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			saveState()
			return
		case <-ticker.C:
			saveState()
		}
	}
}
