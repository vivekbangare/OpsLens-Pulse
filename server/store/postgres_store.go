package store

import (
	"database/sql"
	"time"

	"opslense-pulse/shared"
)

// -------------------------------
// PostgresStore
// -------------------------------
type PostgresStore struct {
	db *sql.DB
}

// -------------------------------
// Constructor
// -------------------------------
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// -------------------------------
// API Keys (Postgres)
// -------------------------------

// CountAPIKeys returns total active keys
func (p *PostgresStore) CountAPIKeys() (int, error) {
	var count int
	err := p.db.QueryRow(`
		SELECT COUNT(*) FROM api_keys
		WHERE is_active = true
	`).Scan(&count)

	return count, err
}

// InsertAPIKey inserts new API key
func (p *PostgresStore) InsertAPIKey(k shared.APIKey) error {

	// tenant_id is UUID in Postgres
	_, err := p.db.Exec(`
		INSERT INTO api_keys
		(id, tenant_id, name, key_hash, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		k.KeyID,
		k.TenantID,
		k.Name,
		k.KeyHash,
		true,
		time.Now(),
	)

	return err
}

// ValidateAPIKey validates raw key and returns tenant_id
func (p *PostgresStore) ValidateAPIKey(rawKey string) (bool, string, error) {

	hash := hashAPIKey(rawKey)

	var tenantID string
	err := p.db.QueryRow(`
		SELECT tenant_id
		FROM api_keys
		WHERE key_hash = $1
		  AND is_active = true
		LIMIT 1
	`, hash).Scan(&tenantID)

	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}

	return true, tenantID, nil
}
