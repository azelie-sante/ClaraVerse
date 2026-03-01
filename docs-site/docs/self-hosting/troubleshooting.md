---
title: Troubleshooting
sidebar_label: Troubleshooting
sidebar_position: 5
---

# Troubleshooting

## First Steps

Before diving into specific issues, always start here:

```bash
# Check which services are running and their health
docker compose ps

# Check backend logs (most issues surface here)
docker compose logs -f backend

# Check all logs at once
docker compose logs -f
```

A healthy deployment shows all services as `healthy` in the `STATUS` column of `docker compose ps`.

## Services Not Starting

### Backend Stuck in "Starting" or "Unhealthy"

The backend depends on MongoDB, MySQL, Redis, and SearXNG all being healthy before it starts. If any dependency fails its health check, the backend will not start.

```bash
# Check which dependency is failing
docker compose ps

# Look at the failing service's logs
docker compose logs mongodb
docker compose logs mysql
docker compose logs redis
docker compose logs searxng
```

**Common causes:**
- MySQL taking too long on first start (migrations running). Wait 30--60 seconds and check again.
- Low memory on the host. MySQL alone needs up to 2 GB.

### Frontend Shows "Unhealthy"

The frontend depends on the backend being healthy. Fix the backend first, then the frontend will follow.

## MongoDB Issues

### Connection Refused

```
error: failed to connect to MongoDB: connection refused
```

**Fix:**

```bash
# Verify MongoDB is running
docker compose ps mongodb

# Check logs
docker compose logs mongodb

# Restart MongoDB
docker compose restart mongodb
```

If using MongoDB Atlas or an external instance, verify `MONGODB_URI` in `.env`:

```
MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/claraverse
```

### Disk Full

MongoDB will stop accepting writes if the volume runs out of space.

```bash
# Check volume usage
docker system df -v | grep mongodb
```

## MySQL Issues

### Migrations Not Running

MySQL migrations are in `./backend/migrations/` and are mounted into `/docker-entrypoint-initdb.d/`. They only run on the **first start** when the data volume is empty.

If you need to re-run migrations:

```bash
# Option 1: Reset MySQL completely (destructive)
docker compose down
docker volume rm claraverseai_mysql-data-new
docker compose up -d

# Option 2: Run migrations manually
docker exec -it claraverse-mysql mysql -u root -p claraverse < backend/migrations/001_init.sql
```

### Authentication Plugin Error

```
Authentication plugin 'caching_sha2_password' cannot be loaded
```

The Compose file already sets `--default-authentication-plugin=mysql_native_password`. If you see this error with an external MySQL instance, configure it to use `mysql_native_password`.

## Redis Issues

### Connection Refused

```bash
# Check Redis health
docker exec -it claraverse-redis redis-cli ping
# Expected response: PONG

# Restart if needed
docker compose restart redis
```

### Memory Limit Reached

Redis is configured with a 100 MB limit and LRU eviction. If you see eviction warnings, increase the limit in `docker-compose.yml`:

```yaml
command: redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
```

## Port Conflicts

If a port is already in use on the host:

```bash
# Find what's using the port
lsof -i :3001
lsof -i :80
lsof -i :27017

# Kill the process
kill -9 <PID>
```

Or change the host port in `docker-compose.yml`:

```yaml
ports:
  - "8080:80"    # Map frontend to 8080 instead of 80
  - "3002:3001"  # Map backend to 3002 instead of 3001
```

## WebSocket Connection Failed

Symptoms: chat messages don't stream; the UI appears stuck after sending a message.

1. **Check CORS** -- Make sure `ALLOWED_ORIGINS` includes the origin the browser is using.
2. **Check WebSocket URL** -- `VITE_WS_URL` must be reachable from the browser (not a Docker-internal address).
3. **Check rate limits** -- Default is 20 WebSocket connections per minute. If exceeded, new connections are rejected.
4. **Check reverse proxy** -- If using Nginx, ensure the `Upgrade` and `Connection` headers are forwarded. See the [Security Guide](./security.md#4-put-a-reverse-proxy-in-front).

## Rate Limiting

If users are being rate-limited unexpectedly:

```bash
# Clear all rate limit keys
docker exec -it claraverse-redis redis-cli KEYS "ratelimit:*"

# Delete a specific key
docker exec -it claraverse-redis redis-cli DEL "ratelimit:<key>"
```

To adjust limits, change the `RATE_LIMIT_*` variables in `.env` and restart the backend:

```bash
docker compose restart backend
```

## Common Error Messages

| Error | Cause | Fix |
|-------|-------|-----|
| `ENCRYPTION_MASTER_KEY is not set` | Missing required env var | Generate with `openssl rand -hex 32` and add to `.env` |
| `JWT_SECRET is not set` | Missing required env var | Generate with `openssl rand -hex 64` and add to `.env` |
| `connection refused` (any DB) | Database container not running or not healthy | Run `docker compose ps` and check the failing service |
| `CORS error` in browser console | Browser origin not in `ALLOWED_ORIGINS` | Add the origin to `ALLOWED_ORIGINS` in `.env` |
| `too many requests` (429) | Rate limit exceeded | Wait, or increase `RATE_LIMIT_*` values |
| `authentication plugin` error (MySQL) | MySQL auth mode mismatch | Use the Compose file's default MySQL config |
| `no space left on device` | Docker volume full | Free disk space or prune unused volumes: `docker system prune --volumes` |

## Checking Logs

```bash
# All services
docker compose logs -f

# Single service with line limit
docker compose logs -f --tail 100 backend

# Filter errors only
docker compose logs backend 2>&1 | grep -E "ERROR|FATAL|panic"
```

You can try enabling debug logging by setting in `.env`:

```
LOG_LEVEL=debug
```

Then restart the backend: `docker compose restart backend`.

:::note
`LOG_LEVEL=debug` may not be supported by all backend versions. If setting this variable has no effect, check the backend logs at the default log level instead.
:::

## Full Reset

If all else fails, reset everything to a clean state:

```bash
# Stop all services and delete all volumes
docker compose down -v

# Remove build cache
docker builder prune -f

# Fresh start
docker compose up -d --build
```

:::caution
`docker compose down -v` deletes all database data, uploaded files, and logs. Back up your volumes first if you have data you want to keep.
:::
