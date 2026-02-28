package store

import (
	"time"
)

type AuditLog struct {
	ID           string
	TenantID     string
	UserID       *string
	Action       string
	ResourceType string
	ResourceID   string
	Status       string
	IPAddress    string
	UserAgent    string
	Metadata     map[string]interface{}
	CreatedAt    time.Time
}
