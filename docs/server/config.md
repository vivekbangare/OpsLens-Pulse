# Server Configuration

## Configuration File Locations

### Linux
```bash 
/etc/opslens-pulse/server-config.yaml
```

### Windows
```powershell
C:\ProgramData\OpsLens-Pulse\server-config.yaml
```
## Example Configuration

```yaml
listen_port: 9898
token: "mysecrettoken"
```

### Environment Variable Override

```bash
export OPS_SERVER_CONFIG=/custom/path/server-config.yaml
```
