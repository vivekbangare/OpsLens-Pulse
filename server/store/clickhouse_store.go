package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"opslense-pulse/shared"
)

type ClickHouseStore struct {
	db *sql.DB
}

/*
|--------------------------------------------------------------------------
| Constructor
|--------------------------------------------------------------------------
*/

func NewClickHouseStore(dsn string) (*ClickHouseStore, error) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(10 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &ClickHouseStore{db: db}, nil
}

/*
|--------------------------------------------------------------------------
| Metrics (time-series, append-only)
|--------------------------------------------------------------------------
*/

func (c *ClickHouseStore) SaveMetrics(m shared.HostMetrics) error {
	_, err := c.db.Exec(`
		INSERT INTO host_metrics
		(
			account_id,
			agent_id,
			hostname,
			ip,
			os,
			cpu_percent,
			mem_used_mb,
			mem_total_mb,
			uptime_sec,
			tags,
			ts
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.AccountID,
		m.AgentID,
		m.Hostname,
		m.IP,
		m.OS,
		m.CPUPercent,
		m.MemUsedMB,
		m.MemTotalMB,
		m.UpTimeSec,
		m.Tags,
		time.Now(),
	)

	return err
}

/*
|--------------------------------------------------------------------------
| Heartbeat (append-only)
|--------------------------------------------------------------------------
*/

func (c *ClickHouseStore) UpdateHeartbeat(hostname string) error {
	_, err := c.db.Exec(`
		INSERT INTO agent_heartbeat
		(hostname, last_seen)
		VALUES (?, ?)
	`,
		hostname,
		time.Now(),
	)

	return err
}

/*
|--------------------------------------------------------------------------
| API Keys
|--------------------------------------------------------------------------
*/

func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

// REQUIRED by Store interface
func (c *ClickHouseStore) CountAPIKeys() (int, error) {
	row := c.db.QueryRow(`SELECT count() FROM api_keys`)

	var count uint64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return int(count), nil
}

// REQUIRED by Store interface
func (c *ClickHouseStore) InsertAPIKey(k APIKey) error {
	_, err := c.db.Exec(`
		INSERT INTO api_keys
		(
			account_id,
			key_id,
			key_hash,
			name,
			is_active,
			is_bootstrap,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		k.AccountID,
		k.KeyID,
		k.KeyHash,
		k.Name,
		k.IsActive,
		k.IsBootstrap,
		time.Now(),
	)

	return err
}

// REQUIRED by Store interface
func (c *ClickHouseStore) ValidateAPIKey(rawKey string) (bool, error) {
	hash := hashAPIKey(rawKey)

	row := c.db.QueryRow(`
		SELECT count()
		FROM api_keys
		WHERE key_hash = ?
		  AND is_active = 1
		LIMIT 1
	`, hash)

	var count uint64
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	if count == 0 {
		return false, errors.New("invalid or inactive api key")
	}

	return true, nil
}
