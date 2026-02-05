# Architecture Overview

OpsLens Pulse follows a server–agent model.

## High-Level Flow
1. Agent starts on target host
2. Agent collects system metrics
3. Agent sends metrics and heartbeat to server
4. Server validates token
5. Server stores data in memory
6. Web UI displays live host status

## Components

### Agent
- Collects CPU, memory, OS, uptime
- Sends heartbeat
- Sends metrics
- Runs on Linux & Windows

### Server
- API layer
- Authentication middleware
- In-memory store
- Web UI

## Communication
- Protocol: HTTP
- Payload: JSON
- Auth: Bearer token
