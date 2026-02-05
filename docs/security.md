# Security Model

OpsLens Pulse uses a simple but effective token-based security model.

## Authentication
- All APIs require `Authorization: Bearer <token>`
- Token is validated on every request
- No anonymous access allowed

## Token Storage
- Tokens are stored only in config files
- Tokens are never logged
- Tokens are never returned in API responses

## Transport Security
⚠️ For production environments:
- Use HTTPS
- Or place behind a reverse proxy (Nginx / ALB / Traefik)

## Future Enhancements
- TLS support
- Role-Based Access Control (RBAC)
