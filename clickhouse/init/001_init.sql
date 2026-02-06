-- database and tables
CREATE DATABASE IF NOT EXISTS opslens;

CREATE TABLE IF NOT EXISTS opslens.hosts
(
    agent_id   String,
    hostname   String,
    account_id String,
    tags       Map(String, String),
    created_at DateTime DEFAULT now()
)
ENGINE = MergeTree
ORDER BY (account_id, agent_id);

CREATE TABLE IF NOT EXISTS opslens.heartbeats
(
    agent_id   String,
    last_seen  DateTime
)
ENGINE = ReplacingMergeTree(last_seen)
ORDER BY agent_id;

CREATE TABLE IF NOT EXISTS opslens.metrics
(
    agent_id     String,
    hostname     String,
    cpu_percent  Float32,
    mem_used_mb  UInt32,
    mem_total_mb UInt32,
    uptime_sec   UInt64,
    ts           DateTime DEFAULT now()
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(ts)
ORDER BY (agent_id, ts);

CREATE TABLE IF NOT EXISTS opslens.api_keys
(
    key_id        String,
    key_hash      String,
    account_id    String,
    name          String,
    is_active     UInt8,
    is_bootstrap  UInt8,
    created_at    DateTime DEFAULT now()
)
ENGINE = MergeTree
ORDER BY (account_id, key_id);
