package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
// Metrics
// -------------------------------
func (c *ClickHouseStore) SaveMetrics(m shared.HostMetrics) error {
	tagsJSON, _ := json.Marshal(m.Tags)
	ttl := m.TTLDays
	if ttl == 0 {
		ttl = 90
	}

	ts := time.Now()
	if m.Timestamp > 0 {
		if m.Timestamp > 1e12 {
			ts = time.UnixMilli(m.Timestamp)
		} else {
			ts = time.Unix(m.Timestamp, 0)
		}
	}

	_, err := c.db.Exec(`
		INSERT INTO metrics
		(account_id, agent_id, hostname, cpu_percent, mem_used_mb, mem_total_mb,
		 disk_used_mb, disk_total_mb, network_in_mb, network_out_mb, uptime_sec, tags, ts, ttl_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.AccountID,
		m.AgentID,
		m.Hostname,
		m.CPUPercent,
		m.MemUsedMB,
		m.MemTotalMB,
		m.DiskUsedMB,
		m.DiskTotalMB,
		m.NetworkInMB,
		m.NetworkOutMB,
		m.UptimeSec,
		string(tagsJSON),
		ts,
		ttl,
	)

	if err != nil {
		return fmt.Errorf("SaveMetrics failed: %w", err)
	}

	// Update heartbeat at the same time
	return c.UpdateHeartbeat(m.AccountID, m.AgentID, m.Hostname)
}

// -------------------------------
// Heartbeat / Agents
// -------------------------------
func (c *ClickHouseStore) UpdateHeartbeat(accountID, agentID, hostname string) error {
	now := time.Now()
	_, err := c.db.Exec(`
		INSERT INTO agents (account_id, agent_id, hostname, last_seen)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE last_seen = ?
	`,
		accountID,
		agentID,
		hostname,
		now,
		now,
	)
	return err
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

func (c *ClickHouseStore) GetLogs(accountID, agentID string, from, to time.Time, level string, limit int) ([]shared.LogEntry, error) {
	query := `
		SELECT agent_id, timestamp, level, message, tags
		FROM logs
		WHERE account_id = ? AND agent_id = ? AND timestamp BETWEEN ? AND ?
	`
	args := []interface{}{accountID, agentID, from, to}

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

		if err := rows.Scan(&le.AgentID, &ts, &le.Level, &le.Message, &tagsJSON); err != nil {
			return nil, err
		}

		le.Tags = map[string]string{}
		json.Unmarshal([]byte(tagsJSON), &le.Tags)
		le.Timestamp = ts.Unix()
		le.HumanTime = ts.Format("2006-01-02 15:04:05")
		logs = append(logs, le)
	}

	return logs, nil
}

// -------------------------------
// Hosts / Metrics with Filters
// -------------------------------
func (c *ClickHouseStore) GetFiltered(accountID string, filters map[string]string) ([]map[string]interface{}, error) {
	query := `
	SELECT agent_id, hostname,
	       any(cpu_percent) AS cpu_percent,
	       any(mem_used_mb) AS mem_used_mb,
	       any(mem_total_mb) AS mem_total_mb,
	       any(disk_used_mb) AS disk_used_mb,
	       any(disk_total_mb) AS disk_total_mb,
	       any(network_in_mb) AS network_in_mb,
	       any(network_out_mb) AS network_out_mb,
	       any(uptime_sec) AS uptime_sec,
	       max(ts) AS last_seen,
	       tags
	FROM metrics
	WHERE account_id = ?
	`
	args := []interface{}{accountID}

	for k, v := range filters {
		query += fmt.Sprintf(" AND JSONExtractString(tags, '%s') = ?", k)
		args = append(args, v)
	}

	query += " GROUP BY agent_id, hostname, tags ORDER BY last_seen DESC"

	rows, err := c.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	out := []map[string]interface{}{}

	for rows.Next() {
		var agentID, hostname, tagsJSON string
		var cpu, memUsed, memTotal, diskUsed, diskTotal, netIn, netOut float32
		var uptime uint64
		var lastSeen time.Time

		if err := rows.Scan(
			&agentID, &hostname,
			&cpu, &memUsed, &memTotal,
			&diskUsed, &diskTotal, &netIn, &netOut,
			&uptime, &lastSeen, &tagsJSON,
		); err != nil {
			return nil, err
		}

		tags := map[string]string{}
		json.Unmarshal([]byte(tagsJSON), &tags)

		out = append(out, map[string]interface{}{
			"agent_id":       agentID,
			"hostname":       hostname,
			"cpu_percent":    cpu,
			"mem_used_mb":    memUsed,
			"mem_total_mb":   memTotal,
			"disk_used_mb":   diskUsed,
			"disk_total_mb":  diskTotal,
			"network_in_mb":  netIn,
			"network_out_mb": netOut,
			"uptime_sec":     uptime,
			"last_seen":      lastSeen,
			"alive":          now.Sub(lastSeen) < 15*time.Second,
			"tags":           tags,
		})
	}

	return out, nil
}

func (c *ClickHouseStore) SaveContainerMetrics(m shared.ContainerMetrics) error {
	ts := time.Now()
	if m.Timestamp > 0 {
		if m.Timestamp > 1e12 {
			ts = time.UnixMilli(m.Timestamp)
		} else {
			ts = time.Unix(m.Timestamp, 0)
		}
	}

	ttl := m.TTLDays
	if ttl == 0 {
		ttl = 90
	}

	_, err := c.db.Exec(`
		INSERT INTO container_metrics
		(account_id, agent_id, hostname, container_id, container_name,
		 cpu_percent, mem_used_mb, mem_total_mb, disk_used_mb, disk_total_mb,
		 network_in_mb, network_out_mb, ts, ttl_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.AccountID, m.AgentID, m.Hostname, m.ContainerID, m.ContainerName,
		m.CPUPercent, m.MemUsedMB, m.MemTotalMB, m.DiskUsedMB, m.DiskTotalMB,
		m.NetworkInMB, m.NetworkOutMB, ts, ttl,
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

// -------------------------------
// Ensure it implements Store interface
// -------------------------------
var _ Store = (*ClickHouseStore)(nil)
