package identity

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func agentIDPath() string {
	if runtime.GOOS == "windows" {
		return `C:\ProgramData\OpsLens-Pulse\agent-id`
	}
	return "/var/lib/opslens-pulse/agent-id"
}

func LoadOrCreateAgentID() (string, error) {
	path := agentIDPath()

	// Check if the file exists
	if data, err := os.ReadFile(path); err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	id := generateRandomID()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(id), 0644); err != nil {
		return "", err
	}
	return id, nil
}

func generateRandomID() string {
	b := make([]byte, 8) // 128-bit ID
	_, _ = rand.Read(b)
	return "agt_" + hex.EncodeToString(b)
}
