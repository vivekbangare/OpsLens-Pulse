USE opslens;

-- =========================================================
-- DAILY USAGE TABLE
-- =========================================================
CREATE TABLE IF NOT EXISTS usage_daily
(
    tenant_id String,
    date Date,

    metrics_points UInt64,
    logs_bytes UInt64,
    container_metrics_points UInt64,
    container_logs_bytes UInt64,

    active_agents UInt32,
    active_containers UInt32
)
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, date);


-- =========================================================
-- HOST METRICS USAGE
-- =========================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS usage_host_metrics_mv
TO usage_daily
AS
SELECT
    tenant_id,
    toDate(timestamp) AS date,
    count() AS metrics_points,
    0 AS logs_bytes,
    0 AS container_metrics_points,
    0 AS container_logs_bytes,
    0 AS active_agents,
    0 AS active_containers
FROM host_metrics
GROUP BY tenant_id, date;


-- =========================================================
-- HOST LOGS USAGE
-- =========================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS usage_logs_mv
TO usage_daily
AS
SELECT
    tenant_id,
    toDate(ts) AS date,
    0 AS metrics_points,
    sum(length(message)) AS logs_bytes,
    0 AS container_metrics_points,
    0 AS container_logs_bytes,
    0 AS active_agents,
    0 AS active_containers
FROM logs
GROUP BY tenant_id, date;


-- =========================================================
-- CONTAINER METRICS USAGE
-- =========================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS usage_container_metrics_mv
TO usage_daily
AS
SELECT
    tenant_id,
    toDate(ts) AS date,

    0 AS metrics_points,
    0 AS logs_bytes,
    count() AS container_metrics_points,
    0 AS container_logs_bytes,

    0 AS active_agents,
    0 AS active_containers

FROM container_metrics
GROUP BY tenant_id, date;


-- =========================================================
-- CONTAINER LOGS USAGE
-- =========================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS usage_container_logs_mv
TO usage_daily
AS
SELECT
    tenant_id,
    toDate(ts) AS date,

    0 AS metrics_points,
    0 AS logs_bytes,
    0 AS container_metrics_points,
    sum(length(message)) AS container_logs_bytes,

    0 AS active_agents,
    0 AS active_containers

FROM container_logs
GROUP BY tenant_id, date;


-- =========================================================
-- ACTIVE AGENTS PER DAY
-- =========================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS usage_active_agents_mv
TO usage_daily
AS
SELECT
    tenant_id,
    toDate(timestamp) AS date,

    0 AS metrics_points,
    0 AS logs_bytes,
    0 AS container_metrics_points,
    0 AS container_logs_bytes,

    countDistinct(agent_id) AS active_agents,
    0 AS active_containers

FROM host_metrics
GROUP BY tenant_id, date;


-- =========================================================
-- ACTIVE CONTAINERS PER DAY
-- =========================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS usage_active_containers_mv
TO usage_daily
AS
SELECT
    tenant_id,
    toDate(ts) AS date,

    0 AS metrics_points,
    0 AS logs_bytes,
    0 AS container_metrics_points,
    0 AS container_logs_bytes,

    0 AS active_agents,
    countDistinct(container_id) AS active_containers

FROM container_metrics
GROUP BY tenant_id, date;