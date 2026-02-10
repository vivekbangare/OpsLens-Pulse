/* =========================================================
   OpsLens Pulse – ClickHouse Init Schema
   ========================================================= */

CREATE DATABASE IF NOT EXISTS opslens;
USE opslens;

-- =========================================================
-- Accounts
-- =========================================================
CREATE TABLE IF NOT EXISTS accounts
(
    account_id String,
    name String,
    created_at DateTime DEFAULT now()
)
ENGINE = MergeTree()
ORDER BY account_id;

-- =========================================================
-- Agents (identity + metadata ONLY)
-- =========================================================
CREATE TABLE IF NOT EXISTS agents
(
    account_id  String,
    agent_id    String,
    hostname    String,
    ip          String,
    os          String,
    version     String,
    environment String,
    tags        String,           -- JSON string
    first_seen  DateTime DEFAULT now()
)
ENGINE = MergeTree()
ORDER BY (account_id, agent_id);

-- =========================================================
-- Agent Heartbeats (ephemeral)
-- =========================================================
CREATE TABLE IF NOT EXISTS agent_heartbeats
(
    account_id String,
    agent_id   String,
    last_seen  DateTime
)
ENGINE = ReplacingMergeTree(last_seen)
ORDER BY (account_id, agent_id);

-- =========================================================
-- Host Metrics (time-series)
-- =========================================================
CREATE TABLE IF NOT EXISTS metrics
(
    account_id     String,
    agent_id       String,
    hostname       String,
    cpu_percent    Float32,
    mem_used_mb    Float32,
    mem_total_mb   Float32,
    disk_used_mb   Float32,
    disk_total_mb  Float32,
    network_in_mb  Float32,
    network_out_mb Float32,
    uptime_sec     UInt64,
    tags           String,
    ts             DateTime,
    ttl_days       UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, ts)
TTL ts + toIntervalDay(ttl_days);

-- =========================================================
-- Host Metrics Aggregates
-- =========================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_aggregates
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, ts)
AS
SELECT
    account_id,
    agent_id,
    hostname,
    ts,
    avg(cpu_percent)  AS avg_cpu,
    max(cpu_percent)  AS max_cpu,
    avg(mem_used_mb)  AS avg_mem_used,
    max(mem_used_mb)  AS max_mem_used
FROM metrics
GROUP BY account_id, agent_id, hostname, ts;

-- =========================================================
-- Host Logs (time-series)
-- =========================================================
CREATE TABLE IF NOT EXISTS logs
(
    account_id String,
    agent_id   String,
    hostname   String,
    timestamp  DateTime, -- event time (from agent)
    ingested_at  DateTime DEFAULT now(), -- server ingest time
    level      String,
    message    String,
    tags       String,
    ttl_days   UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ingested_at)
ORDER BY (account_id, agent_id, ingested_at)
TTL timestamp + toIntervalDay(ttl_days);

-- =========================================================
-- API Keys
-- =========================================================
CREATE TABLE IF NOT EXISTS api_keys
(
    account_id   String,
    key_id       String,
    key_hash     String,
    name         String,
    is_active    UInt8,
    is_bootstrap UInt8,
    created_at   DateTime
)
ENGINE = MergeTree()
ORDER BY (account_id, key_id);

-- =========================================================
-- Events / Alerts
-- =========================================================
CREATE TABLE IF NOT EXISTS events
(
    account_id String,
    agent_id   String,
    hostname   String,
    ts         DateTime DEFAULT now(),
    event_type String,
    severity   String,
    description String,
    tags       String,
    ttl_days   UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, ts)
TTL ts + toIntervalDay(ttl_days);

-- =========================================================
-- Containers (identity)
-- =========================================================
CREATE TABLE IF NOT EXISTS containers
(
    account_id   String,
    agent_id     String,
    container_id String,
    name          String,
    image         String,
    status        String,
    first_seen    DateTime DEFAULT now()
)
ENGINE = MergeTree()
ORDER BY (account_id, agent_id, container_id);

-- =========================================================
-- Container Metrics
-- =========================================================
CREATE TABLE IF NOT EXISTS container_metrics
(
    account_id     String,
    agent_id       String,
    container_id   String,
    name           String,
    cpu_percent    Float32,
    mem_used_mb    Float32,
    mem_total_mb   Float32,
    network_in_mb  Float32,
    network_out_mb Float32,
    ts             DateTime,
    ttl_days       UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, container_id, ts)
TTL ts + toIntervalDay(ttl_days);

-- =========================================================
-- Container Metrics Aggregates
-- =========================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS container_metrics_aggregates
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, container_id, ts)
AS
SELECT
    account_id,
    agent_id,
    container_id,
    name,
    ts,
    avg(cpu_percent) AS avg_cpu,
    max(cpu_percent) AS max_cpu,
    avg(mem_used_mb) AS avg_mem_used,
    max(mem_used_mb) AS max_mem_used
FROM container_metrics
GROUP BY account_id, agent_id, container_id, name, ts;

-- =========================================================
-- Container Logs
-- =========================================================
CREATE TABLE IF NOT EXISTS container_logs
(
    account_id   String,
    agent_id     String,
    container_id String,
    name          String,
    timestamp     DateTime,
    level         String,
    message       String,
    tags          String,
    ttl_days      UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (account_id, agent_id, container_id, timestamp)
TTL timestamp + toIntervalDay(ttl_days);

-- =========================================================
-- Container Events
-- =========================================================
CREATE TABLE IF NOT EXISTS container_events
(
    account_id   String,
    agent_id     String,
    container_id String,
    name          String,
    ts            DateTime DEFAULT now(),
    event_type    String,
    severity      String,
    description   String,
    tags          String,
    ttl_days      UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, container_id, ts)
TTL ts + toIntervalDay(ttl_days);
