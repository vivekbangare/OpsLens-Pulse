# Agent Configuration

## Configuration File Locations

### Linux
`/etc/opslens-pulse/agent-config.yaml`

### Windows
`C:\ProgramData\OpsLens-Pulse\agent-config.yaml`

## Example Configuration

```yaml
server:
  url: "http://localhost:9898"
  token: "mysecrettoken"

agent:
  interval_seconds: 5
  self_upgrade: false

tags:
  env: dev
  role: web
