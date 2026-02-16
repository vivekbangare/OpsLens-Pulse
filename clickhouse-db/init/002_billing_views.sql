USE opslens;

CREATE TABLE IF NOT EXISTS usage_daily
(
    tenant_id String,
    date Date,
    metrics_points UInt64,
    logs_bytes UInt64,
    events_count UInt64,
    active_agents UInt32,
    active_containers UInt32
)
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, date);

-- =============================
-- Metrics usage
-- =============================
CREATE MATERIALIZED VIEW IF NOT EXISTS usage_metrics_mv
TO usage_daily
AS
SELECT
    tenant_id,
    toDate(ts) AS date,
    count() AS metrics_points,
    0 AS logs_bytes,
    0 AS events_count,
    0 AS active_agents,
    0 AS active_containers
FROM metrics
GROUP BY tenant_id, date;

-- =============================
-- Logs usage
-- =============================
CREATE MATERIALIZED VIEW IF NOT EXISTS usage_logs_mv
TO usage_daily
AS
SELECT
    tenant_id,
    toDate(timestamp) AS date,
    0 AS metrics_points,
    sum(length(message)) AS logs_bytes,
    0 AS events_count,
    0 AS active_agents,
    0 AS active_containers
FROM logs
GROUP BY tenant_id, date;
