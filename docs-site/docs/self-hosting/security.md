---
title: Security Guide
sidebar_label: Security
sidebar_position: 4
---

# Security Guide

ClaraVerse uses local JWT authentication, AES-256-GCM encryption, and layered middleware to protect your data. This page covers the security model and best practices for production deployments.

## Authentication

ClaraVerse uses local JWT authentication (no external auth providers required).

### How It Works

1. **Registration** -- Password is hashed with Argon2id and stored in MongoDB.
2. **Login** -- Password is verified against the stored hash. On success, the backend issues an access token and a refresh token (both HS256-signed JWTs).
3. **Authenticated requests** -- The frontend sends the access token in the `Authorization: Bearer <token>` header. Middleware validates the JWT on every request.
4. **Token refresh** -- When the access token expires, the frontend uses the refresh token to get a new pair.

### Token Lifetimes

| Token | Default | Variable |
|-------|---------|----------|
| Access token | 15 minutes | `JWT_ACCESS_TOKEN_EXPIRY` |
| Refresh token | 7 days | `JWT_REFRESH_TOKEN_EXPIRY` |

### Endpoint Security Layers

```
Request
  --> CORS check
    --> Rate limiter
      --> Public endpoints (register, login, health)
      --> JWT validation
        --> User endpoints (chat, agents, workflows)
        --> Admin check
          --> Admin endpoints (user management, analytics)
```

The first registered user automatically becomes admin. Additional superadmins can be designated via `SUPERADMIN_USER_IDS`.

## Encryption

### ENCRYPTION_MASTER_KEY

This is the most critical secret in your deployment. It protects:

- Integration credentials stored in MongoDB
- Optionally synced chat sessions
- Any other user data marked for encryption

**Generate it once and store it securely:**

```bash
openssl rand -hex 32
```

:::danger
If you lose `ENCRYPTION_MASTER_KEY`, all encrypted data becomes permanently unrecoverable. Back up this key separately from your database backups.
:::

### How Encryption Works

1. A single `ENCRYPTION_MASTER_KEY` is set as an environment variable.
2. For each user, a unique per-user key is derived using HKDF with the user's ID as the salt.
3. Data is encrypted with AES-256-GCM, which provides both confidentiality and integrity (authenticated encryption).

```
ENCRYPTION_MASTER_KEY
        |
        v
  HKDF (User ID as salt)
        |
        v
  Per-User Key  -->  AES-256-GCM  -->  Encrypted Data
```

This means:
- Compromising one user's data does not expose other users' data.
- The master key never appears in the database.
- Each encrypted blob includes a GCM authentication tag, so tampering is detected.

## CORS Configuration

The `ALLOWED_ORIGINS` variable controls which origins can make requests to the backend.

**Development (default):**
```
ALLOWED_ORIGINS=http://localhost,http://localhost:5173,http://localhost:5174,http://localhost:3000,http://localhost:8080
```

**Production:**
```
ALLOWED_ORIGINS=https://yourdomain.com
```

Only list the exact origins you need. Do not use `*` in production.

## Rate Limiting

Rate limiting is applied per-IP at the middleware layer. All values are requests per minute.

| Tier | Default | Variable | Applies To |
|------|---------|----------|------------|
| Global API | 200/min | `RATE_LIMIT_GLOBAL_API` | All `/api/*` routes |
| Public read | 120/min | `RATE_LIMIT_PUBLIC_READ` | Unauthenticated GET endpoints |
| Authenticated | 60/min | `RATE_LIMIT_AUTHENTICATED` | Logged-in user requests |
| WebSocket | 20/min | `RATE_LIMIT_WEBSOCKET` | New WebSocket connections |
| Image proxy | 60/min | `RATE_LIMIT_IMAGE_PROXY` | Image proxy requests |

To clear rate limits for a specific user:

```bash
docker exec -it claraverse-redis redis-cli DEL "ratelimit:user_abc123"
```

## Production Best Practices

### 1. Set Environment to Production

```
ENVIRONMENT=production
```

This disables the `DEV_API_KEY` and enables production-level logging.

### 2. Generate Strong Secrets

```bash
# Encryption key (64 hex chars = 256 bits)
openssl rand -hex 32

# JWT secret (128 hex chars = 512 bits)
openssl rand -hex 64
```

Never reuse secrets across environments.

### 3. Change Default Database Passwords

```
MYSQL_ROOT_PASSWORD=<strong-random-password>
MYSQL_PASSWORD=<strong-random-password>
```

### 4. Put a Reverse Proxy in Front

Use Nginx, Caddy, or Traefik as a reverse proxy to:
- Terminate TLS/SSL
- Add HTTP security headers
- Provide an additional layer of rate limiting

Example Nginx snippet:

```nginx
server {
    listen 443 ssl;
    server_name yourdomain.com;

    ssl_certificate     /etc/ssl/certs/yourdomain.pem;
    ssl_certificate_key /etc/ssl/private/yourdomain.key;

    location / {
        proxy_pass http://localhost:80;
    }

    location /api/ {
        proxy_pass http://localhost:3001;
    }

    location /ws/ {
        proxy_pass http://localhost:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### 5. Restrict CORS Origins

Set `ALLOWED_ORIGINS` to your production domain only.

### 6. Lock Down Exposed Ports

In production, consider removing the host port mappings for databases in `docker-compose.yml` so they are only accessible within the Docker network:

```yaml
# Remove these lines for production:
# ports:
#   - "27017:27017"   # MongoDB
#   - "3306:3306"     # MySQL
#   - "6379:6379"     # Redis
```

### 7. Back Up Secrets and Data

- Store `ENCRYPTION_MASTER_KEY` and `JWT_SECRET` in a secrets manager or secure vault.
- Back up MongoDB and MySQL volumes regularly.
- Test your backup restoration process.

### 8. Keep Images Updated

```bash
docker compose pull
docker compose up -d --build
```

Regularly update the base images (`mongo:7`, `mysql:8.0`, `redis:7-alpine`, `searxng/searxng:latest`) to pick up security patches.
