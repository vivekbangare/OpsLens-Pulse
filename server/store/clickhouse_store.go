package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"opslense-pulse/shared"
)

// -------------------------------
// ClickHouseStore struct
// -------------------------------
type ClickHouseStore struct {
	db *sql.DB
}

// -------------------------------
// Constructor
// -------------------------------
func NewClickHouseStore(dsn string) (*ClickHouseStore, error) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(15 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &ClickHouseStore{db: db}, nil
}

// -------------------------------
// API Keys
// -------------------------------
func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

func (c *ClickHouseStore) CountAPIKeys() (int, error) {
	row := c.db.QueryRow(`SELECT count() FROM api_keys`)
	var count uint64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return int(count), nil
}

func (c *ClickHouseStore) InsertAPIKey(k shared.APIKey) error {
	_, err := c.db.Exec(`
		INSERT INTO api_keys
		(account_id, key_id, key_hash, name, is_active, is_bootstrap, created_at)
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

// -------------------------------
// Metrics
// -------------------------------
func (c *ClickHouseStore) SaveMetrics(m shared.HostMetrics) error {
	// Marshal tags once
	tagsJSON, _ := json.Marshal(m.Tags)

	// Resolve timestamp
	ts := time.Now()
	if m.Timestamp > 0 {
		if m.Timestamp > 1e12 {
			ts = time.UnixMilli(m.Timestamp)
		} else {
			ts = time.Unix(m.Timestamp, 0)
		}
	}

	// Resolve TTL
	ttl := m.TTLDays
	if ttl == 0 {
		ttl = 90
	}

	_, err := c.db.Exec(`
		INSERT INTO agents
		(account_id, agent_id, hostname, ip, os, environment, tags, last_seen)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.AccountID,
		m.AgentID,
		m.Hostname,
		m.IP,
		m.OS,
		m.Tags["env"], // environment
		string(tagsJSON),
		ts,
	)

	return err
}


// -------------------------------
// Logs
// -------------------------------
func (c *ClickHouseStore) InsertLogs(batch shared.LogBatch) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO logs
		(account_id, agent_id, hostname, timestamp, level, message, tags, ttl_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, l := range batch.Logs {
		ts := time.Now()
		if l.Timestamp > 0 {
			if l.Timestamp > 1e12 {
				ts = time.UnixMilli(l.Timestamp)
			} else {
				ts = time.Unix(l.Timestamp, 0)
			}
		}

		level := l.Level
		if level == "" {
			level = "info"
		}

		tagsJSON, _ := json.Marshal(l.Tags)
		ttl := l.TTLDays
		if ttl == 0 {
			ttl = 90
		}

		_, err := stmt.Exec(
			batch.AccountID,
			batch.AgentID,
			batch.Hostname,
			ts,
			level,
			l.Message,
			string(tagsJSON),
			ttl,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (c *ClickHouseStore) GetLogs(
	accountID string,
	hostname string,
	agentID string,
	from, to time.Time,
	level string,
	limit int,
) ([]shared.LogEntry, error) {

	query := `
		SELECT agent_id, hostname, timestamp, level, message, tags
		FROM logs
		WHERE account_id = ?
	`
	args := []interface{}{accountID}

	if hostname != "" {
		query += " AND hostname = ?"
		args = append(args, hostname)
	}
	if agentID != "" {
		query += " AND agent_id = ?"
		args = append(args, agentID)
	}
	if !from.IsZero() && !to.IsZero() {
		query += " AND timestamp BETWEEN ? AND ?"
		args = append(args, from, to)
	}
	if level != "" {
		query += " AND level = ?"
		args = append(args, level)
	}

	query += " ORDER BY timestamp ASC LIMIT ?"
	args = append(args, limit)

	rows, err := c.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []shared.LogEntry
	for rows.Next() {
		var le shared.LogEntry
		var ts time.Time
		var tagsJSON string

		if err := rows.Scan(
			&le.AgentID,
			&le.Hostname,
			&ts,
			&le.Level,
			&le.Message,
			&tagsJSON,
		); err != nil {
			return nil, err
		}

		le.Timestamp = ts.Unix()
		le.HumanTime = ts.Format("2006-01-02 15:04:05")
		_ = json.Unmarshal([]byte(tagsJSON), &le.Tags)

		logs = append(logs, le)
	}

	return logs, nil
}

func (c *ClickHouseStore) SaveContainerMetrics(m shared.ContainerMetrics) error {
	ts := time.Now()
	if m.Timestamp > 0 {
		ts = time.Unix(m.Timestamp, 0)
	}

	ttl := m.TTLDays
	if ttl == 0 {
		ttl = 90
	}

	_, err := c.db.Exec(`
		INSERT INTO container_metrics
		(account_id, agent_id, hostname, container_id, name,
		 cpu_percent, mem_used_mb, mem_total_mb,
		 disk_used_mb, disk_total_mb,
		 network_in_mb, network_out_mb,
		 ts, ttl_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.AccountID,
		m.AgentID,
		m.Hostname,
		m.ContainerID,
		m.ContainerName, // ✅ FIXED
		m.CPUPercent,
		m.MemUsedMB,
		m.MemTotalMB,
		m.DiskUsedMB,
		m.DiskTotalMB,
		m.NetworkInMB,
		m.NetworkOutMB,
		ts,
		ttl,
	)

	return err
}


func (c *ClickHouseStore) InsertContainerLogs(batch shared.ContainerLogBatch) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO container_logs
		(account_id, agent_id, hostname, container_id, container_name, timestamp, level, message, tags, ttl_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, l := range batch.Logs {
		ts := time.Now()
		if l.Timestamp > 0 {
			if l.Timestamp > 1e12 {
				ts = time.UnixMilli(l.Timestamp)
			} else {
				ts = time.Unix(l.Timestamp, 0)
			}
		}
		level := l.Level
		if level == "" {
			level = "info"
		}
		tagsJSON, _ := json.Marshal(l.Tags)
		ttl := l.TTLDays
		if ttl == 0 {
			ttl = 90
		}

		_, err := stmt.Exec(
			batch.AccountID, batch.AgentID, batch.Hostname, l.ContainerID, l.ContainerName,
			ts, level, l.Message, string(tagsJSON), ttl,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (c *ClickHouseStore) GetFiltered(
    accountID string,
    resourceType string,
    filters map[string]string,
    start, end time.Time,
    limit int,
) ([]shared.LogEntry, error) {

    if limit <= 0 {
        limit = 100
    }

    query := `
    SELECT
        timestamp,
        level,
        message,
        hostname,
        agent_id
    FROM logs
    WHERE account_id = ?
    `

    args := []any{accountID}

    if !start.IsZero() {
        query += " AND timestamp >= ?"
        args = append(args, start)
    }

    if !end.IsZero() {
        query += " AND timestamp <= ?"
        args = append(args, end)
    }

    for k, v := range filters {
        query += " AND " + k + " = ?"
        args = append(args, v)
    }

    query += " ORDER BY timestamp DESC LIMIT ?"
    args = append(args, limit)

    rows, err := c.db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    logs := []shared.LogEntry{}
    for rows.Next() {
        var l shared.LogEntry
        if err := rows.Scan(
            &l.Timestamp,
            &l.Level,
            &l.Message,
            &l.Hostname,
            &l.AgentID,
        ); err != nil {
            return nil, err
        }
        logs = append(logs, l)
    }

    return logs, nil
}
func (c *ClickHouseStore) UpsertAgentHeartbeat(hb shared.Heartbeat) error {
	_, err := c.db.Exec(`
		INSERT INTO agents
		(account_id, agent_id, hostname, last_seen)
		VALUES (?, ?, ?, ?)
	`,
		hb.AccountID,
		hb.AgentID,
		hb.Hostname,
		hb.Timestamp,
	)
	return err
}

func (c *ClickHouseStore) ListAgents(accountID string) ([]shared.AgentInfo, error) {
	rows, err := c.db.Query(`
		SELECT
			account_id,
			agent_id,
			hostname,
			ip,
			os,
			environment,
			tags,
			first_seen,
			last_seen
		FROM agents
		WHERE account_id = ?
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	aliveWindow := 30 * time.Second

	var out []shared.AgentInfo
	for rows.Next() {
		var a shared.AgentInfo
		var tagsJSON string

		if err := rows.Scan(
			&a.AccountID,
			&a.AgentID,
			&a.Hostname,
			&a.IP,
			&a.OS,
			&a.Environment,
			&tagsJSON,
			&a.FirstSeen,
			&a.LastSeen,
		); err != nil {
			return nil, err
		}

		_ = json.Unmarshal([]byte(tagsJSON), &a.Tags)
		a.Alive = now.Sub(a.LastSeen) <= aliveWindow

		out = append(out, a)
	}

	return out, nil
}


// -------------------------------
// Ensure it implements Store interface
// -------------------------------
var _ Store = (*ClickHouseStore)(nil)
