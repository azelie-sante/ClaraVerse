---
title: Environment Variable Reference
sidebar_label: Environment Variables
sidebar_position: 3
---

# Environment Variable Reference

ClaraVerse uses a single `.env` file in the project root as the source of truth for all configuration. Docker Compose reads it automatically.

```bash
cp .env.example .env
```

## Environment Mode

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `ENVIRONMENT` | `development` | No | Set to `production` for production deployments. Disables dev-only features like `DEV_API_KEY`. |

## Security (Required)

These **must** be set before first start. Generate them with `openssl`:

```bash
echo "ENCRYPTION_MASTER_KEY=$(openssl rand -hex 32)" >> .env
echo "JWT_SECRET=$(openssl rand -hex 64)" >> .env
```

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `ENCRYPTION_MASTER_KEY` | *(empty)* | **Yes** | 256-bit hex key for AES-256-GCM encryption of user data (credentials, chat sync). **Losing this key means losing access to all encrypted data.** |
| `JWT_SECRET` | `change_me_to_a_random_secret_at_least_32_chars` | **Yes** | Secret for signing JWT access and refresh tokens. Minimum 32 characters. |

## JWT Authentication

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `JWT_ACCESS_TOKEN_EXPIRY` | `15m` | No | Access token lifetime. Go duration format (e.g., `15m`, `1h`). |
| `JWT_REFRESH_TOKEN_EXPIRY` | `168h` | No | Refresh token lifetime. Default is 7 days. |

## Frontend Build Args

These are baked into the frontend at Docker build time. If you change them, rebuild the frontend container.

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `VITE_API_BASE_URL` | `http://localhost:3001` | No | Backend API URL as seen by the browser. For production: `https://api.yourdomain.com`. |
| `VITE_WS_URL` | `ws://localhost:3001` | No | WebSocket URL as seen by the browser. For production: `wss://api.yourdomain.com`. |

## Backend Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `ALLOWED_ORIGINS` | `http://localhost,http://localhost:5173,http://localhost:5174,http://localhost:3000,http://localhost:8080` | No | Comma-separated CORS allowed origins. For production, set to your domain only. |
| `FRONTEND_URL` | `http://localhost:5173` | No | Frontend URL used by the backend for redirects and links. |
| `BACKEND_URL` | `http://localhost:3001` | No | Backend public URL used to generate absolute download URLs. |
| `DEV_API_KEY` | `claraverse-dev-key-2024` | No | Development-only API key for testing. Only works when `ENVIRONMENT` is not `production`. **Note:** This is a Docker-internal variable set automatically by `docker-compose.yml`. You do not need to add it to your `.env` file. |

## MySQL Database

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `MYSQL_ROOT_PASSWORD` | `claraverse_root_2024` | No | MySQL root password. Change for production. |
| `MYSQL_PASSWORD` | `claraverse_pass_2024` | No | Password for the `claraverse_user` MySQL account. Change for production. |
| `DATABASE_URL` | *(auto-generated from `MYSQL_PASSWORD`)* | No | Full MySQL connection string. Override to use an external MySQL instance. **Note:** This is a Docker-internal variable set automatically by `docker-compose.yml`. You do not need to add it to your `.env` file unless you are connecting to an external MySQL instance. |

The Docker Compose file auto-creates the database `claraverse` and user `claraverse_user` on first start.

## MongoDB

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `MONGODB_URI` | `mongodb://mongodb:27017/claraverse` | No | MongoDB connection string. For MongoDB Atlas: `mongodb+srv://user:pass@cluster.mongodb.net/claraverse`. |

## Redis

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `REDIS_URL` | `redis://redis:6379` | No | Redis connection string. For Redis Cloud: `redis://user:pass@host:6379`. |

## Rate Limiting

All values are **requests per minute**.

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `RATE_LIMIT_GLOBAL_API` | `200` | No | Global limit for all `/api/*` routes. |
| `RATE_LIMIT_PUBLIC_READ` | `120` | No | Public read-only endpoints. |
| `RATE_LIMIT_AUTHENTICATED` | `60` | No | Authenticated user requests. |
| `RATE_LIMIT_WEBSOCKET` | `20` | No | WebSocket connections per minute. |
| `RATE_LIMIT_IMAGE_PROXY` | `60` | No | Image proxy requests (prevents bandwidth abuse). |

## External Services

### SearXNG

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `SEARXNG_URLS` | `http://searxng:8080` | No | SearXNG instance URL(s). Comma-separated for load balancing. |

### E2B Code Execution

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `E2B_API_KEY` | *(empty)* | No | E2B cloud API key. Get one at [e2b.dev/dashboard](https://e2b.dev/dashboard). Required only if you enable the E2B service. |

### Composio Integrations

All Composio variables are optional. Set them only if you use the corresponding integration.

| Variable | Default | Description |
|----------|---------|-------------|
| `COMPOSIO_API_KEY` | *(empty)* | Master API key from Composio. |
| `COMPOSIO_GOOGLESHEETS_AUTH_CONFIG_ID` | *(empty)* | Google Sheets OAuth config ID. |
| `COMPOSIO_GMAIL_AUTH_CONFIG_ID` | *(empty)* | Gmail OAuth config ID. |
| `COMPOSIO_LINKEDIN_AUTH_CONFIG_ID` | *(empty)* | LinkedIn OAuth config ID. |
| `COMPOSIO_GOOGLECALENDAR_AUTH_CONFIG_ID` | *(empty)* | Google Calendar OAuth config ID. |
| `COMPOSIO_GOOGLEDRIVE_AUTH_CONFIG_ID` | *(empty)* | Google Drive OAuth config ID. |
| `COMPOSIO_CANVA_AUTH_CONFIG_ID` | *(empty)* | Canva OAuth config ID. |
| `COMPOSIO_TWITTER_AUTH_CONFIG_ID` | *(empty)* | Twitter/X OAuth config ID. |
| `COMPOSIO_YOUTUBE_AUTH_CONFIG_ID` | *(empty)* | YouTube OAuth config ID. |
| `COMPOSIO_ZOOM_AUTH_CONFIG_ID` | *(empty)* | Zoom OAuth config ID. |

## Admin

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `SUPERADMIN_USER_IDS` | *(empty)* | No | Comma-separated MongoDB user IDs to grant superadmin role. The first registered user automatically becomes admin. |

## Production Checklist

Before going to production, verify these settings:

```bash
# Required changes
ENVIRONMENT=production
ENCRYPTION_MASTER_KEY=<your-generated-key>
JWT_SECRET=<your-generated-secret>
MYSQL_ROOT_PASSWORD=<strong-password>
MYSQL_PASSWORD=<strong-password>

# Set to your actual domain
VITE_API_BASE_URL=https://api.yourdomain.com
VITE_WS_URL=wss://api.yourdomain.com
ALLOWED_ORIGINS=https://yourdomain.com
FRONTEND_URL=https://yourdomain.com
BACKEND_URL=https://api.yourdomain.com
```
