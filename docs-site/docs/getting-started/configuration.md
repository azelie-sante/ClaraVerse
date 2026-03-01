---
title: Configuration
sidebar_label: Configuration
sidebar_position: 2
---

# Configuration

All configuration is done through the `.env` file in the project root. Docker Compose reads this file automatically. After changing any value, restart the stack:

```bash
docker compose down
docker compose up --build -d
```

## Environment Mode

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `development` | Set to `production` for production deployments. Disables dev-only features like `DEV_API_KEY`. |

## Frontend URLs

These are baked into the frontend at build time. If you change them, you must rebuild the frontend container.

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_BASE_URL` | `http://localhost:3001` | Backend API URL the frontend calls. For production: `https://api.yourdomain.com` |
| `VITE_WS_URL` | `ws://localhost:3001` | WebSocket URL for real-time streaming. For production: `wss://api.yourdomain.com` |

## JWT Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | *(must be set)* | Secret key for signing JWT tokens. Minimum 32 characters. Generate with `openssl rand -hex 32`. |
| `JWT_ACCESS_TOKEN_EXPIRY` | `15m` | How long access tokens are valid. |
| `JWT_REFRESH_TOKEN_EXPIRY` | `168h` | How long refresh tokens are valid (default: 7 days). |

## Backend

| Variable | Default | Description |
|----------|---------|-------------|
| `ALLOWED_ORIGINS` | `http://localhost,http://localhost:5173,http://localhost:5174,http://localhost:3000,http://localhost:8080` | Comma-separated CORS origins. For production: `https://yourdomain.com` |
| `ENCRYPTION_MASTER_KEY` | *(must be set)* | Encrypts user data in MongoDB (conversations, credentials). Generate with `openssl rand -hex 32`. **Do not lose this key.** |
| `FRONTEND_URL` | `http://localhost:5173` | Used by the backend for redirects and email links. |
| `BACKEND_URL` | `http://localhost:3001` | Public URL of the backend, used for generating download links. |

## MySQL Database

| Variable | Default | Description |
|----------|---------|-------------|
| `MYSQL_ROOT_PASSWORD` | `claraverse_root_2024` | MySQL root password. Change for production. |
| `MYSQL_PASSWORD` | `claraverse_pass_2024` | Password for the `claraverse_user` MySQL account. Change for production. |

The MySQL database name (`claraverse`) and user (`claraverse_user`) are set in `docker-compose.yml`.

## MongoDB and Redis

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_URI` | `mongodb://mongodb:27017/claraverse` | MongoDB connection string. For Atlas: `mongodb+srv://user:pass@cluster.mongodb.net/claraverse` |
| `REDIS_URL` | `redis://redis:6379` | Redis connection string. Used for job scheduling and WebSocket pub/sub. |

## External Services

All optional. ClaraVerse works without them.

| Variable | Default | Description |
|----------|---------|-------------|
| `E2B_API_KEY` | *(empty)* | API key for [E2B](https://e2b.dev) code execution sandbox. Enables the code interpreter skill. |
| `SEARXNG_URLS` | `http://searxng:8080` | Internal SearXNG URL(s). Comma-separated for load balancing. Pre-configured in Docker Compose. |
| `COMPOSIO_API_KEY` | *(empty)* | API key for [Composio](https://composio.dev) integrations (Google Sheets, Gmail, etc.). |
| `COMPOSIO_GOOGLESHEETS_AUTH_CONFIG_ID` | *(empty)* | Composio auth config for Google Sheets integration. |
| `COMPOSIO_GMAIL_AUTH_CONFIG_ID` | *(empty)* | Composio auth config for Gmail integration. |
| `COMPOSIO_LINKEDIN_AUTH_CONFIG_ID` | *(empty)* | Composio auth config for LinkedIn integration. |
| `COMPOSIO_GOOGLECALENDAR_AUTH_CONFIG_ID` | *(empty)* | Composio auth config for Google Calendar integration. |
| `COMPOSIO_GOOGLEDRIVE_AUTH_CONFIG_ID` | *(empty)* | Composio auth config for Google Drive integration. |
| `COMPOSIO_CANVA_AUTH_CONFIG_ID` | *(empty)* | Composio auth config for Canva integration. |
| `COMPOSIO_TWITTER_AUTH_CONFIG_ID` | *(empty)* | Composio auth config for Twitter/X integration. |
| `COMPOSIO_YOUTUBE_AUTH_CONFIG_ID` | *(empty)* | Composio auth config for YouTube integration. |
| `COMPOSIO_ZOOM_AUTH_CONFIG_ID` | *(empty)* | Composio auth config for Zoom integration. |

## Admin

| Variable | Default | Description |
|----------|---------|-------------|
| `SUPERADMIN_USER_IDS` | *(empty)* | Comma-separated user IDs to grant superadmin access. The first registered user is automatically admin; use this for additional admins. |

## Rate Limiting

All values are requests per minute.

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_GLOBAL_API` | `200` | Global limit for all `/api/*` routes. |
| `RATE_LIMIT_PUBLIC_READ` | `120` | Limit for public read-only endpoints. |
| `RATE_LIMIT_AUTHENTICATED` | `60` | Limit for authenticated user requests. |
| `RATE_LIMIT_WEBSOCKET` | `20` | Limit for new WebSocket connections. |
| `RATE_LIMIT_IMAGE_PROXY` | `60` | Limit for image proxy requests (prevents bandwidth abuse). |

## Production Checklist

When deploying to production, at minimum:

1. Set `ENVIRONMENT=production`.
2. Generate strong, unique values for `JWT_SECRET` and `ENCRYPTION_MASTER_KEY`.
3. Change `MYSQL_ROOT_PASSWORD` and `MYSQL_PASSWORD` from defaults.
4. Set `VITE_API_BASE_URL` and `VITE_WS_URL` to your public domain (with `https://` and `wss://`).
5. Set `ALLOWED_ORIGINS` to your frontend domain only.
6. Set `FRONTEND_URL` and `BACKEND_URL` to your public URLs.
7. Place a reverse proxy (nginx, Caddy, Traefik) in front to handle TLS.
