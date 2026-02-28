package store

import (
	"context"
	"encoding/json"
	"opslense-pulse/shared"
	"time"
)

////////////////////////////////////////////////////////////
// GetLogs
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) GetLogs(
	ctx context.Context,
	tenantID string,
	hostname string,
	agentID string,
	from, to time.Time,
	level string,
	source string,
	sourceType string,
	limit int,
) ([]shared.LogEntry, error) {

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := `
		SELECT
			agent_id,
			hostname,
			source_name,
			source_type,
			ts,
			level,
			message
		FROM logs
		WHERE tenant_id = ?
	`

	args := []any{tenantID}

	if source != "" {
		query += " AND source_name = ?"
		args = append(args, source)
	}

	if sourceType != "" {
		query += " AND source_type = ?"
		args = append(args, sourceType)
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
		query += " AND ts >= ?"
		args = append(args, from)
	}

	if !to.IsZero() {
		query += " AND ts <= ?"
		args = append(args, to)
	}

	if level != "" {
		query += " AND level = ?"
		args = append(args, level)
	}

	query += " ORDER BY ts DESC LIMIT ?"
	args = append(args, limit)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []shared.LogEntry

	for rows.Next() {
		var le shared.LogEntry
		var ts time.Time

		if err := rows.Scan(
			&le.AgentID,
			&le.Hostname,
			&le.SourceName,
			&le.SourceType,
			&ts,
			&le.Level,
			&le.Message,
		); err != nil {
			return nil, err
		}

		le.TenantID = tenantID
		le.Timestamp = ts.Unix()
		le.HumanTime = ts.Format("2006-01-02 15:04:05")

		logs = append(logs, le)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

////////////////////////////////////////////////////////////
// ListAgents
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) ListAgents(
	ctx context.Context,
	tenantID string,
) ([]shared.AgentInfo, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			a.tenant_id,
			a.agent_id,
			any(a.hostname) AS hostname,
			any(a.ip) AS ip,
			any(a.public_ip) AS public_ip,
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
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []shared.AgentInfo

	for rows.Next() {
		var ai shared.AgentInfo
		var tagsJSON string

		if err := rows.Scan(
			&ai.TenantID,
			&ai.AgentID,
			&ai.Hostname,
			&ai.IP,
			&ai.PublicIP,
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

		if !ai.LastSeen.IsZero() {
			ai.Alive = time.Since(ai.LastSeen) <= 10*time.Second
		}

		result = append(result, ai)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

////////////////////////////////////////////////////////////
// GetLatestHostMetrics
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) GetLatestHostMetrics(
	ctx context.Context,
	tenantID string,
) (map[string]shared.HostMetrics, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			agent_id,
			argMax(hostname, timestamp),
			argMax(cpu_percent, timestamp),
			argMax(mem_used_mb, timestamp),
			argMax(mem_total_mb, timestamp),
			argMax(uptime_sec, timestamp)
		FROM host_metrics
		WHERE tenant_id = ?
		GROUP BY agent_id
	`, tenantID)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

////////////////////////////////////////////////////////////
// GetLogSources
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) GetLogSources(
	ctx context.Context,
	tenantID, agentID string,
) ([]string, error) {

	rows, err := c.db.QueryContext(ctx, `
		SELECT DISTINCT JSONExtractString(tags, 'source')
		FROM logs
		WHERE tenant_id = ?
		  AND agent_id = ?
		  AND JSONHas(tags, 'source')
	`, tenantID, agentID)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}

////////////////////////////////////////////////////////////
// SearchLogs (Unified View)
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) SearchLogs(
	ctx context.Context,
	tenantID string,
	req shared.LogSearchRequest,
) ([]map[string]interface{}, error) {

	fromTime, err := time.Parse(time.RFC3339, req.From)
	if err != nil {
		return nil, err
	}

	toTime, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT tenant_id, agent_id, source_name, source_type,
		       ts, level, message
		FROM unified_logs
		WHERE tenant_id = ?
		  AND ts BETWEEN ? AND ?
	`

	args := []any{tenantID, fromTime, toTime}

	if req.Level != "" {
		query += " AND level = ?"
		args = append(args, req.Level)
	}

	if req.Source != "" {
		query += " AND source_name = ?"
		args = append(args, req.Source)
	}

	if req.Query != "" {
		query += " AND positionCaseInsensitive(message, ?) > 0"
		args = append(args, req.Query)
	}

	query += " ORDER BY ts DESC"

	if req.Limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, req.Limit, req.Offset)
	} else {
		query += " LIMIT 500"
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}

	for rows.Next() {
		var tid, agentID, sourceName, sourceType string
		var ts time.Time
		var level, message string

		if err := rows.Scan(
			&tid,
			&agentID,
			&sourceName,
			&sourceType,
			&ts,
			&level,
			&message,
		); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"tenant_id":   tid,
			"agent_id":    agentID,
			"source_name": sourceName,
			"source_type": sourceType,
			"ts":          ts.Unix(),
			"level":       level,
			"message":     message,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (c *ClickHouseStore) GetLogsTimeline(
	ctx context.Context,
	tenantID string,
	agentID string,
	sourceType string,
	from, to time.Time,
) (TimelineResult, error) {

	query := `
		SELECT
			bucket,
			source_type,
			total,
			errors,
			warnings,
			if(total = 0,
			   0,
			   toFloat64(errors) / toFloat64(total)
			) AS error_rate
		FROM
		(
			SELECT
				bucket,
				source_type,
				sum(total)    AS total,
				sum(errors)   AS errors,
				sum(warnings) AS warnings
			FROM logs_minute_agg
			WHERE tenant_id = ?
			  AND bucket >= ?
			  AND bucket <= ?
	`

	args := []any{tenantID, from, to}

	if agentID != "" {
		query += " AND agent_id = ?"
		args = append(args, agentID)
	}

	if sourceType != "" {
		query += " AND source_type = ?"
		args = append(args, sourceType)
	}

	query += `
			GROUP BY bucket, agent_id, source_type
		)
		ORDER BY bucket ASC
	`
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return TimelineResult{}, err
	}
	defer rows.Close()

	var points []TimelinePoint

	for rows.Next() {
		var p TimelinePoint

		if err := rows.Scan(
			&p.Bucket,
			&p.SourceType,
			&p.Total,
			&p.Errors,
			&p.Warnings,
			&p.ErrorRate,
		); err != nil {
			return TimelineResult{}, err
		}

		points = append(points, p)
	}

	if err := rows.Err(); err != nil {
		return TimelineResult{}, err
	}

	anomaly := DetectZScore(points)

	return TimelineResult{
		Points:  points,
		Anomaly: anomaly,
	}, nil
}
