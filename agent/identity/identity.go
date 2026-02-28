package identity

import (
	"crypto/rand"
	"encoding/hex"
	"opslense-pulse/agent/collector"
	"os"
	"path/filepath"
	"strings"
)

func LoadOrCreateAgentID() (string, error) {
	path := filepath.Join(collector.StateDir, "agent-id")

	// Check if the file exists
	if data, err := os.ReadFile(path); err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	id := generateRandomID()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(id), 0600); err != nil {
		return "", err
	}
	return id, nil
}

func generateRandomID() string {
	b := make([]byte, 16) // 128-bit ID
	_, err := rand.Read(b)
	if err != nil {
		panic("crypto rand failed")
	}
	return "agt_" + hex.EncodeToString(b)
}
