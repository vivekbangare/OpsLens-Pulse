1. Server & API Readiness
✅ Ensure all API endpoints are covered:
/api/heartbeat
/api/metrics
/api/logs
/api/logs/fetch
/api/hosts
✅ Authentication works for all APIs (Authorization: Bearer <API_KEY>).
✅ Graceful shutdown implemented (SIGTERM, Ctrl+C).
2. Configuration & Secrets
✅ Server config file exists and validated (server-config.yaml).
✅ ClickHouse DSN / credentials secure and not hard-coded.
✅ API bootstrap key stored securely; do not expose in logs after first run.
✅ Optional environment variables for sensitive configs (e.g., OPS_SERVER_CONFIG).
3. Database & Persistence
✅ ClickHouse schema is fully applied:
metrics (time-series)
logs (time-series)
agents and agent_tags
api_keys
events (optional)
✅ TTLs configured for metrics/logs (metrics.ttl_days, logs.ttl_days).
✅ Materialized views set up for metrics aggregation.
✅ Check database connectivity & retries in server logs.
4. Logging & Observability
✅ Server logs persist to file or stdout/stderr (consider log rotation).
✅ Structured logging for easier parsing (JSON format optional).
✅ Monitor server logs for errors, e.g., failed API key validation, DB errors.
✅ Optional: integrate Prometheus or ELK for server metrics/logs.
5. Security
✅ Use HTTPS for all traffic in production (reverse proxy like Nginx / Traefik recommended).
✅ All APIs require a valid API key.
✅ Avoid exposing raw server errors in API responses.
✅ Secure ClickHouse with user/password; restrict IP access.
✅ Limit permissions for agents to write-only where possible.
6. High Availability / Scaling
✅ Ensure ClickHouse cluster or replication for fault tolerance if needed.
✅ Consider running multiple server instances behind a load balancer.
✅ If high-frequency metrics are expected, validate ClickHouse write throughput.
7. Monitoring & Alerts
✅ Heartbeat monitoring:
Mark hosts as alive or offline based on last_seen.
✅ Track API response times and errors.
✅ Alert on DB connectivity issues, server downtime, or high metrics ingestion lag.
8. Backup & Disaster Recovery
✅ Periodic ClickHouse backups for metrics/logs/agents.
✅ Store backup offsite (S3, NFS, or cloud storage).
✅ Test restore process from backup.
9. Performance Optimization
✅ Indexing in ClickHouse is correct (ORDER BY, PARTITION BY).
✅ Use batch inserts for logs & metrics to reduce DB load.
✅ RWMutex used in memory store — ok for dev, but production relies on DB.
10. Optional Enhancements
✅ Rate limiting per agent to prevent overload.
✅ Metrics aggregation for dashboards (materialized views already in place).
✅ Tag-based filtering for multi-environment visibility.

| ✅ Task                 | Details / Verification                                         |
| ---------------------- | -------------------------------------------------------------- |
| **Server Config**      | `server-config.yaml` exists, valid, ports correct              |
| **ClickHouse DSN**     | Connection works; credentials secure; TLS if possible          |
| **Bootstrap API Key**  | Generated once; stored securely; no exposure in logs           |
| **APIs Auth**          | All endpoints require `Authorization: Bearer <API_KEY>`        |
| **Metrics Table**      | `metrics` exists, partitions & TTL set                         |
| **Logs Table**         | `logs` exists, partitions & TTL set                            |
| **Agents Table**       | `agents` & `agent_tags` exist                                  |
| **API Keys Table**     | `api_keys` exists and works for validation                     |
| **Materialized Views** | `metrics_aggregates` exists for dashboards                     |
| **Logging**            | Server logs enabled; rotation or stdout/stderr handled         |
| **HTTPS / Security**   | Reverse proxy or TLS; no plain HTTP                            |
| **Graceful Shutdown**  | Server handles `SIGTERM` & `Ctrl+C`                            |
| **Heartbeat**          | Agents’ `last_seen` updated correctly                          |
| **High Availability**  | Optional: multiple server instances / load balancer            |
| **Backup**             | ClickHouse backups tested; restore procedure documented        |
| **Monitoring**         | Server metrics, agent heartbeats, API response times monitored |
| **Rate Limits**        | Optional: ensure no agent overload                             |
| **Tag-based Filters**  | Environment / tags filtering working as expected               |
| **Staging Test**       | Fully tested in staging environment before production          |


AGENT:

. Suggested Production Checklist for Agent
| ✅ Task                 | Verification / Notes                                                                  |
| ---------------------- | ------------------------------------------------------------------------------------- |
| **Config YAML**        | Exists, `server.url` & `server.api_key` correct.                                      |
| **Agent ID**           | Persistent, file exists at `/var/lib/opslens-pulse/agent-id` (Linux) or Windows path. |
| **Logs Path**          | Log file path configured and readable.                                                |
| **Metrics Collection** | CPU, memory, host info verified.                                                      |
| **Heartbeat**          | Heartbeat reaches server at configured interval.                                      |
| **Network**            | IP detection works; firewall allows outbound to server.                               |
| **Retries**            | Sending metrics/logs/heartbeat retry logic works.                                     |
| **Logging**            | Errors logged to stdout/file; rotation if needed.                                     |
| **Graceful Shutdown**  | Handles SIGTERM/SIGINT cleanly.                                                       |
| **Docker Metrics**     | Docker installed or collector handles missing Docker gracefully.                      |
| **SelfUpgrade**        | If enabled, upgrades tested in staging.                                               |
| **Tagging**            | Tags applied correctly in metrics & logs.                                             |
| **Security**           | API key stored securely; HTTPS recommended for production.                            |
| **Resource Usage**     | Agent CPU/memory minimal, no memory leaks.                                            |
| **Monitoring**         | Server verifies last heartbeat and log arrivals.                                      |
