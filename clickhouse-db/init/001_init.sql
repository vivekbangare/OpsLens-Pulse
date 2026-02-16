-- =========================================================
-- CREATE DATABASE opslens
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
    os LowCardinality(String),
    version String,
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
-- HOST METRICS (RAW)
-- =========================================================
CREATE TABLE IF NOT EXISTS metrics
(
    tenant_id String,
    agent_id String,
    hostname LowCardinality(String),

    cpu_percent Float32,
    mem_used_mb Float32,
    mem_total_mb Float32,
    disk_used_mb Float32,
    disk_total_mb Float32,
    network_in_mb Float32,
    network_out_mb Float32,
    uptime_sec UInt64,

    ts DateTime,
    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (tenant_id, agent_id, ts)
TTL ts + toIntervalDay(ttl_days)
SETTINGS index_granularity = 8192;

-- =========================================================
-- HOST METRICS AGGREGATION (1m)
-- =========================================================
CREATE TABLE IF NOT EXISTS metrics_agg_1m
(
    tenant_id String,
    agent_id String,
    hostname LowCardinality(String),
    bucket_start DateTime,

    cpu_avg_state AggregateFunction(avg, Float32),
    cpu_max_state AggregateFunction(max, Float32),
    mem_avg_state AggregateFunction(avg, Float32),
    mem_max_state AggregateFunction(max, Float32),
    disk_avg_state AggregateFunction(avg, Float32),
    disk_max_state AggregateFunction(max, Float32)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(bucket_start)
ORDER BY (tenant_id, agent_id, bucket_start);

CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_agg_1m_mv
TO metrics_agg_1m
AS
SELECT
    tenant_id,
    agent_id,
    hostname,
    toStartOfMinute(ts) AS bucket_start,

    avgState(cpu_percent)  AS cpu_avg_state,
    maxState(cpu_percent)  AS cpu_max_state,

    avgState(mem_used_mb)  AS mem_avg_state,
    maxState(mem_used_mb)  AS mem_max_state,

    avgState(disk_used_mb) AS disk_avg_state,
    maxState(disk_used_mb) AS disk_max_state

FROM metrics
GROUP BY tenant_id, agent_id, hostname, bucket_start;

-- =========================================================
-- CONTAINER METRICS (RAW)
-- =========================================================
CREATE TABLE IF NOT EXISTS container_metrics
(
    tenant_id String,
    agent_id String,
    container_id String,
    name LowCardinality(String),

    cpu_percent Float32,
    mem_used_mb Float32,
    mem_total_mb Float32,
    disk_used_mb Float32,
    disk_total_mb Float32,
    network_in_mb Float32,
    network_out_mb Float32,
    ts DateTime,
    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (tenant_id, agent_id, container_id, ts)
TTL ts + toIntervalDay(ttl_days);

-- =========================================================
-- CONTAINER METRICS AGGREGATION (1m)
-- =========================================================
CREATE TABLE IF NOT EXISTS container_metrics_agg_1m
(
    tenant_id String,
    agent_id String,
    container_id String,
    name LowCardinality(String),
    bucket_start DateTime,

    cpu_avg_state AggregateFunction(avg, Float32),
    cpu_max_state AggregateFunction(max, Float32),
    mem_avg_state AggregateFunction(avg, Float32),
    mem_max_state AggregateFunction(max, Float32),
    disk_avg_state AggregateFunction(avg, Float32),
    disk_max_state AggregateFunction(max, Float32)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(bucket_start)
ORDER BY (tenant_id, agent_id, container_id, bucket_start);

CREATE MATERIALIZED VIEW IF NOT EXISTS container_metrics_agg_1m_mv
TO container_metrics_agg_1m
AS
SELECT
    tenant_id,
    agent_id,
    container_id,
    name,
    toStartOfMinute(ts) AS bucket_start,

    avgState(cpu_percent)  AS cpu_avg_state,
    maxState(cpu_percent)  AS cpu_max_state,

    avgState(mem_used_mb)  AS mem_avg_state,
    maxState(mem_used_mb)  AS mem_max_state,

    avgState(disk_used_mb) AS disk_avg_state,
    maxState(disk_used_mb) AS disk_max_state

FROM container_metrics
GROUP BY tenant_id, agent_id, container_id, name, bucket_start;

-- =========================================================
-- HOST LOGS
-- =========================================================
CREATE TABLE IF NOT EXISTS logs
(
    tenant_id String,
    agent_id String,
    hostname LowCardinality(String),
    timestamp DateTime,
    level LowCardinality(String),
    message String CODEC(ZSTD),
    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, agent_id, timestamp)
TTL timestamp + toIntervalDay(ttl_days);

-- =========================================================
-- CONTAINER LOGS
-- =========================================================
CREATE TABLE IF NOT EXISTS container_logs
(
    tenant_id String,
    agent_id String,
    container_id String,
    name LowCardinality(String),
    timestamp DateTime,
    level LowCardinality(String),
    message String CODEC(ZSTD),
    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, agent_id, container_id, timestamp)
TTL timestamp + toIntervalDay(ttl_days);

-- =========================================================
-- HOST EVENTS
-- =========================================================
CREATE TABLE IF NOT EXISTS events
(
    tenant_id String,
    agent_id String,
    hostname LowCardinality(String),
    ts DateTime DEFAULT now(),
    event_type LowCardinality(String),
    severity LowCardinality(String),
    description String CODEC(ZSTD),
    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (tenant_id, agent_id, ts)
TTL ts + toIntervalDay(ttl_days);

-- =========================================================
-- CONTAINER EVENTS
-- =========================================================
CREATE TABLE IF NOT EXISTS container_events
(
    tenant_id String,
    agent_id String,
    container_id String,
    name LowCardinality(String),
    ts DateTime DEFAULT now(),
    event_type LowCardinality(String),
    severity LowCardinality(String),
    description String CODEC(ZSTD),
    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (tenant_id, agent_id, container_id, ts)
TTL ts + toIntervalDay(ttl_days);
