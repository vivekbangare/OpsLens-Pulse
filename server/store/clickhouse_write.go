package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"opslense-pulse/shared"
	"strings"
)

////////////////////////////////////////////////////////////
// SaveMetrics
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) SaveMetrics(
	ctx context.Context,
	m shared.HostMetrics,
) error {

	ts := resolveTimestamp(m.Timestamp)

	// ---------------- host_metrics (single row) ----------------
	if err := c.execWithRetry(ctx, `
		INSERT INTO host_metrics
		(tenant_id, agent_id, hostname, os, version,
		timestamp, cores,
		cpu_percent, cpu_critical, cpu_spike,
		mem_used_mb, mem_total_mb, mem_critical, mem_pressure,
		uptime_sec,
		agent_cpu_percent, agent_mem_mb,
		agent_goroutines, agent_uptime_sec,
		metrics_failures, log_failures)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.TenantID,
		m.AgentID,
		m.Hostname,
		m.OS,
		m.Version,
		ts,
		m.Cores,
		m.CPUPercent,
		boolToUInt8(m.CPUCritical),
		boolToUInt8(m.CPUSpike),
		m.MemUsedMB,
		m.MemTotalMB,
		boolToUInt8(m.MemCritical),
		boolToUInt8(m.MemPressure),
		m.UptimeSec,
		m.Agent.CPUPercent,
		m.Agent.MemoryMB,
		m.Agent.Goroutines,
		m.Agent.UptimeSec,
		m.Agent.MetricsFailures,
		m.Agent.LogFailures,
	); err != nil {
		return err
	}

	// ---------------- host_disk_metrics ----------------
	if len(m.Disks) > 0 {

		diskChunks := chunkLogs(m.Disks, maxBatchSize)

		for _, part := range diskChunks {

			query := `
				INSERT INTO host_disk_metrics
				(tenant_id, agent_id, hostname,
				 timestamp,
				 mount_point, fs_type,
				 total_mb, used_mb, used_pct,
				 spike_detected, critical)
				VALUES
			`

			valueStrings := make([]string, 0, len(part))
			args := make([]interface{}, 0, len(part)*11)

			for _, d := range part {

				valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

				args = append(args,
					m.TenantID,
					m.AgentID,
					m.Hostname,
					ts,
					d.MountPoint,
					d.FSType,
					d.TotalMB,
					d.UsedMB,
					d.UsedPct,
					boolToUInt8(d.SpikeDetected),
					boolToUInt8(d.Critical),
				)
			}

			query += strings.Join(valueStrings, ",")

			if err := c.execWithRetry(ctx, query, args...); err != nil {
				return err
			}
		}
	}

	// ---------------- host_network_interfaces ----------------
	if len(m.Network) > 0 {

		netChunks := chunkLogs(m.Network, maxBatchSize)

		for _, part := range netChunks {

			query := `
				INSERT INTO host_network_interfaces
				(tenant_id, agent_id, hostname,
				 timestamp,
				 interface, in_bps, out_bps)
				VALUES
			`

			valueStrings := make([]string, 0, len(part))
			args := make([]interface{}, 0, len(part)*7)

			for _, n := range part {

				valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?)")

				args = append(args,
					m.TenantID,
					m.AgentID,
					m.Hostname,
					ts,
					n.Name,
					n.InBps,
					n.OutBps,
				)
			}

			query += strings.Join(valueStrings, ",")

			if err := c.execWithRetry(ctx, query, args...); err != nil {
				return err
			}
		}
	}

	return nil
}

////////////////////////////////////////////////////////////
// InsertLogs (Host Logs)
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) InsertLogs(
	ctx context.Context,
	batch shared.LogBatch,
) error {

	if len(batch.Logs) == 0 {
		return nil
	}

	chunks := chunkLogs(batch.Logs, maxBatchSize)

	for _, part := range chunks {

		query := `
			INSERT INTO logs
			(tenant_id, agent_id, hostname,
			 source_name, source_type,
			 ts, level, message, tags, ttl_days)
			VALUES
		`

		valueStrings := make([]string, 0, len(part))
		args := make([]interface{}, 0, len(part)*10)

		for _, l := range part {

			ts := resolveTimestamp(l.Timestamp)

			level := l.Level
			if level == "" {
				level = "info"
			}

			sourceName := l.SourceName
			if sourceName == "" {
				sourceName = "unknown"
			}

			sourceType := l.SourceType
			if sourceType == "" {
				sourceType = "host"
			}

			tagsJSON, _ := json.Marshal(l.Tags)

			ttl := l.TTLDays
			if ttl == 0 {
				ttl = 90
			}

			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

			args = append(args,
				batch.TenantID,
				batch.AgentID,
				batch.Hostname,
				sourceName,
				sourceType,
				ts,
				level,
				l.Message,
				string(tagsJSON),
				ttl,
			)
		}

		query += strings.Join(valueStrings, ",")

		if err := c.execWithRetry(ctx, query, args...); err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////
// InsertContainerLogs
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) InsertContainerLogs(
	ctx context.Context,
	batch shared.ContainerLogBatch,
) error {

	if len(batch.Logs) == 0 {
		return nil
	}

	chunks := chunkLogs(batch.Logs, maxBatchSize)

	for _, part := range chunks {

		query := `
			INSERT INTO container_logs
			(tenant_id, agent_id, container_id,
			 name, ts, level, message, ttl_days)
			VALUES
		`

		valueStrings := make([]string, 0, len(part))
		args := make([]interface{}, 0, len(part)*8)

		for _, l := range part {

			ts := resolveTimestamp(l.Timestamp)

			level := l.Level
			if level == "" {
				level = "info"
			}

			ttl := l.TTLDays
			if ttl == 0 {
				ttl = 90
			}

			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?)")

			args = append(args,
				batch.TenantID,
				batch.AgentID,
				l.ContainerID,
				l.ContainerName,
				ts,
				level,
				l.Message,
				ttl,
			)
		}

		query += strings.Join(valueStrings, ",")

		if err := c.execWithRetry(ctx, query, args...); err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////
// SaveContainerMetrics
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) SaveContainerMetrics(
	ctx context.Context,
	m shared.ContainerMetrics,
) error {

	ts := resolveTimestamp(m.Timestamp)

	ttl := m.TTLDays
	if ttl == 0 {
		ttl = 90
	}

	return c.execWithRetry(ctx, `
		INSERT INTO container_metrics
		(tenant_id, agent_id, hostname,
		container_id, name, image, status,
		cpu_percent, mem_used_mb, mem_total_mb,
		ts, ttl_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.TenantID,
		m.AgentID,
		m.Hostname,
		m.ContainerID,
		m.ContainerName,
		m.Image,
		m.Status,
		m.CPUPercent,
		m.MemUsedMB,
		m.MemTotalMB,
		ts,
		ttl,
	)
}

////////////////////////////////////////////////////////////
// UpsertAgentMetadata
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) UpsertAgentMetadata(
	ctx context.Context,
	m shared.HostMetrics,
) error {

	tagsJSON, _ := json.Marshal(m.Tags)

	var exists int
	err := c.db.QueryRowContext(ctx, `
		SELECT 1
		FROM agents
		WHERE tenant_id = ? AND agent_id = ?
		LIMIT 1
	`,
		m.TenantID,
		m.AgentID,
	).Scan(&exists)

	if err == sql.ErrNoRows {

		return c.execWithRetry(ctx, `
			INSERT INTO agents
			(tenant_id, agent_id, hostname,
			private_ip, public_ip, remote_ip, k8s_node_ip,
			os, version, environment, tags)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			m.TenantID,
			m.AgentID,
			m.Hostname,
			m.PrivateIP,
			m.PublicIP,
			m.RemoteIP,
			m.K8sNodeIP,
			m.OS,
			m.Version,
			m.Tags["env"],
			string(tagsJSON),
		)
	}

	return err
}

////////////////////////////////////////////////////////////
// UpsertAgentHeartbeat
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) UpsertAgentHeartbeat(
	ctx context.Context,
	hb shared.Heartbeat,
) error {

	ts := resolveTimestamp(hb.Timestamp)

	return c.execWithRetry(ctx, `
		INSERT INTO agent_heartbeats
		(tenant_id, agent_id, last_seen)
		VALUES (?, ?, ?)
	`,
		hb.TenantID,
		hb.AgentID,
		ts,
	)
}
