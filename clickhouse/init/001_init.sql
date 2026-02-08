CREATE DATABASE IF NOT EXISTS opslens;

-- Use the opslens database
USE opslens;

-- =====================================
-- Accounts Table
-- =====================================
CREATE TABLE IF NOT EXISTS accounts
(
    account_id String,
    name String,
    created_at DateTime DEFAULT now(),
    PRIMARY KEY (account_id)
) ENGINE = MergeTree()
ORDER BY account_id;

-- =====================================
-- Agents Table
-- =====================================
CREATE TABLE IF NOT EXISTS agents
(
    agent_id String,
    account_id String,
    hostname String,
    ip String,
    os String,
    version String,
    environment String,
    first_seen DateTime DEFAULT now(),
    last_seen DateTime,
    PRIMARY KEY (account_id, agent_id)
) ENGINE = ReplacingMergeTree(last_seen)
ORDER BY (account_id, agent_id);

-- =====================================
-- Agent Tags Table
-- =====================================
CREATE TABLE IF NOT EXISTS agent_tags
(
    agent_id String,
    account_id String,
    tag_key String,
    tag_value String,
    PRIMARY KEY (account_id, agent_id, tag_key)
) ENGINE = MergeTree()
ORDER BY (account_id, agent_id, tag_key);

-- =====================================
-- Metrics Table (time-series)
-- =====================================
CREATE TABLE IF NOT EXISTS metrics (
    account_id String,
    agent_id String,
    hostname String,
    cpu_percent Float32,
    mem_used_mb Float32,
    mem_total_mb Float32,
    disk_used_mb Float32,
    disk_total_mb Float32,
    network_in_mb Float32,
    network_out_mb Float32,
    uptime_sec UInt64,
    tags String,
    ts DateTime,
    ttl_days UInt16 DEFAULT 90
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, ts)
TTL ts + toIntervalDay(ttl_days)
SETTINGS index_granularity = 8192;

-- =====================================
-- Metrics Aggregates (Materialized View)
-- =====================================
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_aggregates
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, ts) AS
SELECT
    account_id,
    agent_id,
    hostname,
    ts,
    avg(cpu_percent) AS avg_cpu,
    max(cpu_percent) AS max_cpu,
    avg(mem_used_mb) AS avg_mem_used,
    max(mem_used_mb) AS max_mem_used
FROM metrics
GROUP BY account_id, agent_id, hostname, ts;

-- =====================================
-- Logs Table (time-series)
-- =====================================
CREATE TABLE IF NOT EXISTS logs (
    account_id String,
    agent_id String,
    hostname String,
    timestamp DateTime,
    level String,
    message String,
    tags String,
    ttl_days UInt16 DEFAULT 90
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (account_id, agent_id, timestamp)
TTL timestamp + toIntervalDay(ttl_days)
SETTINGS index_granularity = 8192;

-- =====================================
-- API Keys Table
-- =====================================
CREATE TABLE IF NOT EXISTS api_keys (
    account_id String,
    key_id String,
    key_hash String,
    name String,
    is_active UInt8,
    is_bootstrap UInt8,
    created_at DateTime
) ENGINE = MergeTree()
ORDER BY (account_id, key_id);

-- =====================================
-- Events / Alerts Table (optional)
-- =====================================
CREATE TABLE IF NOT EXISTS events
(
    account_id String,
    agent_id String,
    hostname String,
    ts DateTime DEFAULT now(),
    event_type String,
    severity String,
    description String,
    tags Nested (
        key String,
        value String
    ),
    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, ts)
TTL ts + toIntervalDay(ttl_days);


-- =====================================
-- Containers Table
-- =====================================

CREATE TABLE IF NOT EXISTS containers
(
    container_id String,
    agent_id String,
    account_id String,
    name String,
    image String,
    status String,
    created_at DateTime DEFAULT now(),
    started_at DateTime,
    stopped_at DateTime,
    PRIMARY KEY (account_id, agent_id, container_id)
) ENGINE = ReplacingMergeTree(started_at)
ORDER BY (account_id, agent_id, container_id);

-- =====================================
-- Container Metrics Table
-- =====================================

CREATE TABLE IF NOT EXISTS container_metrics (
    account_id String,
    agent_id String,
    container_id String,
    name String,
    cpu_percent Float32,
    mem_used_mb Float32,
    mem_total_mb Float32,
    network_in_mb Float32,
    network_out_mb Float32,
    disk_used_mb Float32,
    disk_total_mb Float32,
    uptime_sec UInt64,
    tags String,
    ts DateTime,
    ttl_days UInt16 DEFAULT 90
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, container_id, ts)
TTL ts + toIntervalDay(ttl_days)
SETTINGS index_granularity = 8192;

-- =====================================
-- Container Logs Table
-- =====================================
CREATE TABLE IF NOT EXISTS container_logs (
    account_id String,
    agent_id String,
    container_id String,
    name String,
    timestamp DateTime,
    level String,
    message String,
    tags String,
    ttl_days UInt16 DEFAULT 90
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (account_id, agent_id, container_id, timestamp)
TTL timestamp + toIntervalDay(ttl_days)
SETTINGS index_granularity = 8192;


-- =====================================
-- Container Events Table
-- =====================================
CREATE TABLE IF NOT EXISTS container_events
(
    account_id String,
    agent_id String,
    container_id String,
    name String,
    ts DateTime DEFAULT now(),
    event_type String,
    severity String,
    description String,
    tags Nested (key String, value String),
    ttl_days UInt16 DEFAULT 90
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, container_id, ts)
TTL ts + toIntervalDay(ttl_days);

-- =====================================
-- Container Tags Table
-- =====================================
CREATE TABLE IF NOT EXISTS container_tags
(
    container_id String,
    agent_id String,
    account_id String,
    tag_key String,
    tag_value String,
    PRIMARY KEY (account_id, agent_id, container_id, tag_key)
) ENGINE = MergeTree()
ORDER BY (account_id, agent_id, container_id, tag_key);


-- =====================================
-- Container Metrics Aggregates (Materialized View)
-- =====================================

CREATE MATERIALIZED VIEW IF NOT EXISTS container_metrics_aggregates
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (account_id, agent_id, container_id, ts) AS
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
