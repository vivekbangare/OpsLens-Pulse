package store

import "time"

type APIKey struct {
	AccountID   string
	KeyID       string
	KeyHash     string
	Name        string
	IsActive    bool
	IsBootstrap bool
	CreatedAt   time.Time
}
