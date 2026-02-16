package store

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"opslense-pulse/shared"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

// -------------------------------
// ClickHouseStore
// -------------------------------
type ClickHouseStore struct {
	db *sql.DB
}

// -------------------------------
// Constructor
// -------------------------------
func NewClickHouseStore(dsn string) (*ClickHouseStore, error) {
	log.Println("🧪 ClickHouse DSN:", dsn)
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
// Host Metrics (time-series ONLY)
// -------------------------------
func (c *ClickHouseStore) SaveMetrics(m shared.HostMetrics) error {
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
		(tenant_id, agent_id, hostname,
		 cpu_percent, mem_used_mb, mem_total_mb,
		 disk_used_mb, disk_total_mb,
		 network_in_mb, network_out_mb,
		 uptime_sec, ts, ttl_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.TenantID,
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
		ts,
		ttl,
	)

	return err
}

// -------------------------------
// Logs Query
// -------------------------------
func (c *ClickHouseStore) GetLogs(
	TenantID string,
	hostname string,
	agentID string,
	from, to time.Time,
	level string,
	source string,
	limit int,
) ([]shared.LogEntry, error) {

	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT
			agent_id,
			hostname,
			timestamp,
			level,
			message
		FROM logs
		WHERE tenant_id = ?
	`
	args := []any{TenantID}

	if source != "" {
		query += " AND JSONExtractString(tags, 'source') = ?"
		args = append(args, source)
	}
	if agentID != "" {
		query += " AND agent_id = ?"
		args = append(args, agentID)
	}
	if hostname != "" {
		query += " AND hostname = ?"
		args = append(args, hostname)
	}
	if !from.IsZero() {
		query += " AND timestamp >= ?"
		args = append(args, from)
	}
	if !to.IsZero() {
		query += " AND timestamp <= ?"
		args = append(args, to)
	}
	if level != "" {
		query += " AND level = ?"
		args = append(args, level)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := c.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]shared.LogEntry, 0)

	for rows.Next() {
		var le shared.LogEntry
		var ts time.Time

		if err := rows.Scan(
			&le.AgentID,
			&le.Hostname,
			&ts,
			&le.Level,
			&le.Message,
		); err != nil {
			continue
		}

		le.TenantID = TenantID
		le.Timestamp = ts.Unix()
		le.HumanTime = ts.Format("2006-01-02 15:04:05")

		logs = append(logs, le)
	}

	if err := rows.Err(); err != nil {
		return logs, err
	}

	return logs, nil
}

// -------------------------------
// Agent Metadata (identity + tags)
// -------------------------------
func (c *ClickHouseStore) UpsertAgentMetadata(m shared.HostMetrics) error {
	tagsJSON, _ := json.Marshal(m.Tags)

	_, err := c.db.Exec(`
		INSERT INTO agents
		(tenant_id, agent_id, hostname, ip, os, version, environment, tags, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.TenantID,
		m.AgentID,
		m.Hostname,
		m.IP,
		m.OS,
		m.Version,
		m.Tags["env"],
		string(tagsJSON),
		time.Now(),
	)

	return err
}

// -------------------------------
// Agent Heartbeat
// -------------------------------
func (c *ClickHouseStore) UpsertAgentHeartbeat(hb shared.Heartbeat) error {
	_, err := c.db.Exec(`
		INSERT INTO agent_heartbeats
		(tenant_id, agent_id, last_seen)
		VALUES (?, ?, ?)
	`,
		hb.TenantID,
		hb.AgentID,
		hb.Timestamp,
	)
	return err
}

// -------------------------------
// Hosts Listing (JOIN + alive)
// -------------------------------
func (c *ClickHouseStore) ListAgents(TenantID string) ([]shared.AgentInfo, error) {
	rows, err := c.db.Query(`
		SELECT
			a.tenant_id,
			a.agent_id,
			any(a.hostname) AS hostname,
			any(a.ip) AS ip,
			any(a.os) AS os,
			any(a.version) AS version,
			any(a.environment) AS environment,
			any(a.tags) AS tags,
			min(a.first_seen) AS first_seen,
			max(h.last_seen) AS last_seen
		FROM agents a
		LEFT JOIN agent_heartbeats h
		  ON a.tenant_id = h.tenant_id
		 AND a.agent_id = h.agent_id
		WHERE a.tenant_id = ?
		GROUP BY a.tenant_id, a.agent_id
	`, TenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []shared.AgentInfo
	for rows.Next() {
		var ai shared.AgentInfo
		var tagsJSON string

		if err := rows.Scan(
			&ai.TenantID,
			&ai.AgentID,
			&ai.Hostname,
			&ai.IP,
			&ai.OS,
			&ai.Version,
			&ai.Environment,
			&tagsJSON,
			&ai.FirstSeen,
			&ai.LastSeen,
		); err != nil {
			return nil, err
		}

		_ = json.Unmarshal([]byte(tagsJSON), &ai.Tags)
		ai.Alive = time.Since(ai.LastSeen) <= 10*time.Second

		out = append(out, ai)
	}

	return out, nil
}

// -------------------------------
// Container Logs
// -------------------------------
func (c *ClickHouseStore) InsertContainerLogs(batch shared.ContainerLogBatch) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO container_logs
		(tenant_id, agent_id, hostname, container_id, name,
		 timestamp, level, message, tags, ttl_days)
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
		log.Printf(
			"🧪 CH INSERT logs: account=%s agent=%s host=%s ts=%v",
			batch.TenantID,
			batch.AgentID,
			batch.Hostname,
			ts,
		)
		_, err := stmt.Exec(
			batch.TenantID,
			batch.AgentID,
			batch.Hostname,
			l.ContainerID,
			l.ContainerName,
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

// -------------------------------
// Host Logs
// -------------------------------
func (c *ClickHouseStore) InsertLogs(batch shared.LogBatch) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO logs
		(tenant_id, agent_id, hostname, timestamp, level, message, ttl_days)
		VALUES (?, ?, ?, ?, ?, ?, ?)
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

		ttl := l.TTLDays
		if ttl == 0 {
			ttl = 90
		}

		_, err := stmt.Exec(
			batch.TenantID,
			batch.AgentID,
			batch.Hostname,
			ts,
			level,
			l.Message,
			ttl,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// -------------------------------
// Container Metrics
// -------------------------------
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
		(tenant_id, agent_id, container_id, name,
		 cpu_percent, mem_used_mb, mem_total_mb,
		 disk_used_mb, disk_total_mb,
		 network_in_mb, network_out_mb, ts, ttl_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.TenantID,
		m.AgentID,
		m.ContainerID,
		m.ContainerName,
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

func (c *ClickHouseStore) GetLatestHostMetrics(TenantID string) (map[string]shared.HostMetrics, error) {
	rows, err := c.db.Query(`
		SELECT
			agent_id,
			hostname,
			argMax(cpu_percent, ts),
			argMax(mem_used_mb, ts),
			argMax(mem_total_mb, ts),
			argMax(uptime_sec, ts)
		FROM metrics
		WHERE tenant_id = ?
		GROUP BY agent_id, hostname
	`, TenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]shared.HostMetrics)

	for rows.Next() {
		var m shared.HostMetrics
		if err := rows.Scan(
			&m.AgentID,
			&m.Hostname,
			&m.CPUPercent,
			&m.MemUsedMB,
			&m.MemTotalMB,
			&m.UptimeSec,
		); err != nil {
			return nil, err
		}
		out[m.AgentID] = m
	}

	return out, nil
}

func (c *ClickHouseStore) GetLogSources(
	TenantID, agentID string,
) ([]string, error) {

	rows, err := c.db.Query(`
		SELECT DISTINCT JSONExtractString(tags, 'source')
		FROM logs
		WHERE tenant_id = ?
		  AND agent_id = ?
		  AND JSONHas(tags, 'source')
	`, TenantID, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []string
	for rows.Next() {
		var src string
		if err := rows.Scan(&src); err == nil && src != "" {
			sources = append(sources, src)
		}
	}
	return sources, nil
}

// -------------------------------
// Interface check
// -------------------------------
var _ Store = (*ClickHouseStore)(nil)
