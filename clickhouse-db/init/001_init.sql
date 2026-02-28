-- =========================================================
-- DATABASE
-- =========================================================
CREATE DATABASE IF NOT EXISTS opslens;
USE opslens;

-- =========================================================
-- AGENTS
-- =========================================================
CREATE TABLE IF NOT EXISTS agents
(
    tenant_id String,
    agent_id String,

    hostname LowCardinality(String),
    ip String,
    public_ip String,

    os LowCardinality(String),
    version LowCardinality(String),
    environment LowCardinality(String),

    tags String,

    first_seen DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (tenant_id, agent_id);


-- =========================================================
-- HEARTBEATS
-- =========================================================
CREATE TABLE IF NOT EXISTS agent_heartbeats
(
    tenant_id String,
    agent_id String,
    last_seen DateTime
)
ENGINE = ReplacingMergeTree(last_seen)
ORDER BY (tenant_id, agent_id)
TTL last_seen + INTERVAL 30 DAY;


-- =========================================================
-- HOST METRICS (ANOMALY-AWARE)
-- =========================================================
CREATE TABLE IF NOT EXISTS host_metrics
(
    tenant_id String,
    agent_id String,
    hostname LowCardinality(String),
    os LowCardinality(String),
    version LowCardinality(String),

    timestamp DateTime,

    cores UInt16,
    cpu_percent Float32,
    cpu_critical UInt8,
    cpu_spike UInt8,

    mem_used_mb Float32,
    mem_total_mb Float32,
    mem_critical UInt8,
    mem_pressure UInt8,

    uptime_sec UInt64,

    ip String,
    public_ip String,

    -- Agent internal health
    agent_cpu_percent Float32,
    agent_mem_mb Float32,
    agent_goroutines UInt32,
    agent_uptime_sec UInt64,
    metrics_failures UInt64,
    log_failures UInt64,

    tags String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, agent_id, timestamp)
TTL timestamp + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;


-- =========================================================
-- HOST DISK METRICS (MULTI-MOUNT SUPPORT)
-- =========================================================
CREATE TABLE IF NOT EXISTS host_disk_metrics
(
    tenant_id String,
    agent_id String,
    hostname LowCardinality(String),

    timestamp DateTime,

    mount_point String,
    fs_type LowCardinality(String),

    total_mb Float32,
    used_mb Float32,
    used_pct Float32,

    spike_detected UInt8,
    critical UInt8
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, agent_id, mount_point, timestamp)
TTL timestamp + INTERVAL 30 DAY;


-- =========================================================
-- HOST NETWORK INTERFACES
-- =========================================================
CREATE TABLE IF NOT EXISTS host_network_interfaces
(
    tenant_id String,
    agent_id String,
    hostname LowCardinality(String),

    timestamp DateTime,

    interface LowCardinality(String),
    in_bps Float32,
    out_bps Float32
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, agent_id, interface, timestamp)
TTL timestamp + INTERVAL 30 DAY;


-- =========================================================
-- HOST METRICS AGGREGATION (1 MINUTE)
-- =========================================================
CREATE TABLE IF NOT EXISTS host_metrics_agg_1m
(
    tenant_id String,
    agent_id String,
    hostname LowCardinality(String),
    bucket_start DateTime,

    cpu_avg_state AggregateFunction(avg, Float32),
    cpu_max_state AggregateFunction(max, Float32),

    mem_avg_state AggregateFunction(avg, Float32),
    mem_max_state AggregateFunction(max, Float32)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(bucket_start)
ORDER BY (tenant_id, agent_id, bucket_start);

CREATE MATERIALIZED VIEW IF NOT EXISTS host_metrics_agg_1m_mv
TO host_metrics_agg_1m
AS
SELECT
    tenant_id,
    agent_id,
    hostname,
    toStartOfMinute(timestamp) AS bucket_start,

    avgState(cpu_percent)  AS cpu_avg_state,
    maxState(cpu_percent)  AS cpu_max_state,

    avgState(mem_used_mb)  AS mem_avg_state,
    maxState(mem_used_mb)  AS mem_max_state

FROM host_metrics
GROUP BY tenant_id, agent_id, hostname, bucket_start;


-- =========================================================
-- CONTAINER METRICS
-- =========================================================
CREATE TABLE IF NOT EXISTS container_metrics
(
    tenant_id String,
    agent_id String,
    hostname LowCardinality(String),

    container_id String,
    name LowCardinality(String),
    image LowCardinality(String),
    status LowCardinality(String),
    cpu_percent Float32,
    mem_used_mb Float32,
    mem_total_mb Float32,

    ts DateTime,
    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (tenant_id, agent_id, container_id, ts)
TTL ts + toIntervalDay(ttl_days);


-- =========================================================
-- HOST LOGS
-- =========================================================
CREATE TABLE IF NOT EXISTS logs
(
    tenant_id String,
    agent_id String,
    hostname LowCardinality(String),

    source_name String,
    source_type LowCardinality(String),

    ts DateTime64(3),
    level LowCardinality(String),
    message String CODEC(ZSTD),

    tags String,

    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (tenant_id, agent_id, ts)
TTL ts + toIntervalDay(ttl_days)
SETTINGS index_granularity = 8192;


-- =========================================================
-- CONTAINER LOGS
-- =========================================================
CREATE TABLE IF NOT EXISTS container_logs
(
    tenant_id String,
    agent_id String,
    container_id String,
    name LowCardinality(String),

    ts DateTime64(3),
    level LowCardinality(String),
    message String CODEC(ZSTD),

    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (tenant_id, agent_id, container_id, ts)
TTL ts + toIntervalDay(ttl_days)
SETTINGS index_granularity = 8192;


-- =========================================================
-- UNIFIED LOG VIEW
-- =========================================================
CREATE VIEW IF NOT EXISTS unified_logs AS
SELECT
    tenant_id,
    agent_id,
    source_name,
    source_type,
    ts,
    level,
    message
FROM logs
UNION ALL
SELECT
    tenant_id,
    agent_id,
    name AS source_name,
    'container' AS source_type,
    ts,
    level,
    message
FROM container_logs;


-- =========================================================
-- LOGS MINUTE AGGREGATION
-- =========================================================
CREATE TABLE IF NOT EXISTS logs_minute_agg
(
    tenant_id String,
    agent_id String,
    source_type LowCardinality(String),
    bucket DateTime,

    total UInt64,
    errors UInt64,
    warnings UInt64
)
ENGINE = SummingMergeTree
PARTITION BY toYYYYMM(bucket)
ORDER BY (tenant_id, agent_id, source_type, bucket);


CREATE MATERIALIZED VIEW IF NOT EXISTS logs_minute_mv_host
TO logs_minute_agg
AS
SELECT
    tenant_id,
    agent_id,
    source_type,
    toStartOfMinute(ts) AS bucket,
    count() AS total,
    countIf(level = 'error') AS errors,
    countIf(level = 'warn') AS warnings
FROM logs
GROUP BY tenant_id, agent_id, source_type, bucket;


CREATE MATERIALIZED VIEW IF NOT EXISTS logs_minute_mv_container
TO logs_minute_agg
AS
SELECT
    tenant_id,
    agent_id,
    'container' AS source_type,
    toStartOfMinute(ts) AS bucket,
    count() AS total,
    countIf(level = 'error') AS errors,
    countIf(level = 'warn') AS warnings
FROM container_logs
GROUP BY tenant_id, agent_id, bucket;