---
title: Docker Compose Guide
sidebar_label: Docker Compose
sidebar_position: 2
---

# Docker Compose Guide

ClaraVerse ships a single `docker-compose.yml` that starts all six services. This page covers every service, its resource limits, volumes, health checks, and how to customize the deployment.

## Quick Start

```bash
git clone https://github.com/claraverse-space/ClaraVerse.git
cd ClaraVerse
cp .env.example .env

# Generate required secrets
echo "ENCRYPTION_MASTER_KEY=$(openssl rand -hex 32)" >> .env
echo "JWT_SECRET=$(openssl rand -hex 64)" >> .env

# Start all services
docker compose up -d --build
```

The frontend is available at `http://localhost` and the backend API at `http://localhost:3001`. The first user to register becomes admin.

## Service Details

### Frontend

| Setting | Value |
|---------|-------|
| Image | Built from `./frontend/Dockerfile` |
| Container | `claraverse-frontend` |
| Port | `80:80` |
| Depends on | Backend (healthy) |
| Health check | `wget http://127.0.0.1/health` every 30s |

The frontend is a production Vite build served by Nginx. Build-time args `VITE_API_BASE_URL` and `VITE_WS_URL` are baked into the bundle at build.

### Backend

| Setting | Value |
|---------|-------|
| Image | Built from `./backend/Dockerfile` |
| Container | `claraverse-backend` |
| Port | `3001:3001` |
| Memory limit | 1 GB |
| CPU limit | 2.0 |
| Depends on | MongoDB, MySQL, Redis, SearXNG (all healthy) |
| Health check | `wget http://localhost:3001/health` every 30s |

**Volumes:**

| Volume | Mount Point | Purpose |
|--------|-------------|---------|
| `backend-data` | `/app/data` | Persistent application data |
| `backend-uploads` | `/app/uploads` | User file uploads |
| `backend-logs` | `/app/logs` | Application logs |

### MongoDB

| Setting | Value |
|---------|-------|
| Image | `mongo:7` |
| Container | `claraverse-mongodb` |
| Port | `27017:27017` |
| Memory limit | 1 GB |
| CPU limit | 1.0 |
| Health check | `mongosh --eval "db.adminCommand('ping')"` every 30s |

**Volume:** `mongodb-data` mounted at `/data/db`

### MySQL

| Setting | Value |
|---------|-------|
| Image | `mysql:8.0` |
| Container | `claraverse-mysql` |
| Port | `3306:3306` |
| Memory limit | 2 GB |
| CPU limit | 1.0 |
| Health check | `mysqladmin ping` every 10s |
| Character set | `utf8mb4` / `utf8mb4_unicode_ci` |

**Volumes:**

| Volume | Mount Point | Purpose |
|--------|-------------|---------|
| `mysql-data-new` | `/var/lib/mysql` | Database files |
| `./backend/migrations` | `/docker-entrypoint-initdb.d` (read-only) | Auto-run migrations on first start |

### Redis

| Setting | Value |
|---------|-------|
| Image | `redis:7-alpine` |
| Container | `claraverse-redis` |
| Port | `6379:6379` |
| Max memory | 100 MB (LRU eviction) |
| Health check | `redis-cli ping` every 10s |
| Persistence | AOF (append-only file) |

**Volume:** `redis-data` mounted at `/data`

### SearXNG

| Setting | Value |
|---------|-------|
| Image | `searxng/searxng:latest` |
| Container | `claraverse-searxng` |
| Port | Internal only (8080) |
| Health check | `wget http://localhost:8080/` every 30s |
| Config | `./searxng/settings.yml` mounted read-only |

SearXNG is not exposed to the host by default. The backend accesses it over the internal Docker network at `http://searxng:8080`.

## Port Mapping Reference

| Host Port | Container | Service |
|-----------|-----------|---------|
| 80 | frontend:80 | Web UI (Nginx) |
| 3001 | backend:3001 | API + WebSocket |
| 27017 | mongodb:27017 | MongoDB |
| 3306 | mysql:3306 | MySQL |
| 6379 | redis:6379 | Redis |

To change a host port, edit the left side of the mapping in `docker-compose.yml`:

```yaml
ports:
  - "8080:80"   # Frontend on port 8080 instead of 80
```

## Volumes

All named volumes use the `local` driver by default:

| Volume | Service | Purpose |
|--------|---------|---------|
| `backend-data` | Backend | Application data |
| `backend-uploads` | Backend | File uploads |
| `backend-logs` | Backend | Log files |
| `mongodb-data` | MongoDB | Database files |
| `mysql-data-new` | MySQL | Database files |
| `redis-data` | Redis | AOF persistence |

To back up a volume:

```bash
docker run --rm -v claraverseai_mongodb-data:/data -v $(pwd):/backup \
  alpine tar czf /backup/mongodb-backup.tar.gz -C /data .
```

## Enabling the E2B Service (Optional)

E2B provides sandboxed Python code execution. To enable it:

1. Get an API key from [e2b.dev/dashboard](https://e2b.dev/dashboard).
2. Set `E2B_API_KEY` in your `.env` file.
3. Uncomment the `e2b-service` block in `docker-compose.yml`:

```yaml
e2b-service:
  build:
    context: ./backend/e2b-service
    dockerfile: Dockerfile
  container_name: claraverse-e2b
  environment:
    - E2B_API_KEY=${E2B_API_KEY}
  restart: unless-stopped
  healthcheck:
    test: ["CMD", "python", "-c", "import requests; requests.get('http://localhost:8001/health')"]
    interval: 30s
    timeout: 10s
    start_period: 5s
    retries: 3
  networks:
    - claraverse-network
```

The backend automatically connects to `http://e2b-service:8001`.

## Common Operations

### Rebuild a single service

```bash
docker compose up -d --build backend
```

### View logs

```bash
docker compose logs -f backend
docker compose logs -f --tail 100 mysql
```

### Restart all services

```bash
docker compose restart
```

### Stop everything

```bash
docker compose down
```

### Reset all data (destructive)

```bash
docker compose down -v
docker compose up -d --build
```

### Check service health

```bash
docker compose ps
```

The `STATUS` column shows `healthy`, `starting`, or `unhealthy` for each service.

## Resource Requirements

Minimum recommended for a single-server deployment:

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 2 cores | 4 cores |
| RAM | 4 GB | 8 GB |
| Disk | 10 GB | 20 GB+ |

The total memory limits configured in `docker-compose.yml` sum to ~4 GB (MongoDB 1 GB + MySQL 2 GB + Backend 1 GB + Redis 100 MB). The remaining RAM is used by the frontend Nginx process, SearXNG, and the OS.
