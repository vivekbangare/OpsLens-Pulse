package store

import (
	"crypto/sha256"
	"encoding/hex"
)

func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}
