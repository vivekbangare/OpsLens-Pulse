# Server API Reference

## Authentication Header
```http
Authorization: Bearer <token>
```

## Endpoints

```bash
POST /api/metrics
```
Receives system metrics from agent.

```bash
POST /api/heartbeat
```
Receives heartbeat signal from agent.

```bash
GET /api/hosts
```
Returns list of all registered hosts.