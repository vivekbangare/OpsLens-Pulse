package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"opslense-pulse/shared"

	"github.com/google/uuid"
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

	hash := HashAPIKey(rawKey)

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

func (p *PostgresStore) InsertAuditLog(ctx context.Context, log AuditLog) error {

	metadataJSON, _ := json.Marshal(log.Metadata)

	_, err := p.db.ExecContext(ctx, `
		INSERT INTO audit_logs
		(id, tenant_id, user_id, action,
		 resource_type, resource_id,
		 status, ip_address, user_agent,
		 metadata, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`,
		uuid.NewString(),
		log.TenantID,
		log.UserID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.Status,
		log.IPAddress,
		log.UserAgent,
		metadataJSON,
		time.Now(),
	)

	return err
}

func (p *PostgresStore) GetAuditLogs(
	ctx context.Context,
	tenantID string,
	from, to time.Time,
	userID string,
	action string,
	status string,
	limit int,
) ([]AuditLog, error) {

	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query := `
		SELECT id, tenant_id, user_id,
		       action, resource_type, resource_id,
		       status, ip_address, user_agent,
		       metadata, created_at
		FROM audit_logs
		WHERE tenant_id = $1
		  AND created_at BETWEEN $2 AND $3
	`

	args := []interface{}{tenantID, from, to}
	argPos := 4

	if userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, userID)
		argPos++
	}

	if action != "" {
		query += fmt.Sprintf(" AND action = $%d", argPos)
		args = append(args, action)
		argPos++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argPos)
	args = append(args, limit)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog

	for rows.Next() {
		var l AuditLog
		var metadataBytes []byte

		err := rows.Scan(
			&l.ID,
			&l.TenantID,
			&l.UserID,
			&l.Action,
			&l.ResourceType,
			&l.ResourceID,
			&l.Status,
			&l.IPAddress,
			&l.UserAgent,
			&metadataBytes,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal(metadataBytes, &l.Metadata)

		logs = append(logs, l)
	}

	return logs, rows.Err()
}
