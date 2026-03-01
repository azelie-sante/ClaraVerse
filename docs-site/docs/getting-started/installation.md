---
title: Installation
sidebar_label: Installation
sidebar_position: 1
---

# Installation

Get ClaraVerse running with Docker Compose. Total time: about 5 minutes.

## Prerequisites

| Requirement | Minimum | Notes |
|-------------|---------|-------|
| **Docker** | 20.10+ | [Install Docker](https://docs.docker.com/get-docker/) |
| **Docker Compose** | v2+ | Included with Docker Desktop; on Linux, install the `docker-compose-plugin` |
| **Git** | Any recent version | To clone the repository |
| **RAM** | 4 GB free | MongoDB, MySQL, Redis, SearXNG, backend, and frontend all run simultaneously |
| **Disk** | 5 GB free | For Docker images and database volumes |

## Step 1: Clone the Repository

```bash
git clone https://github.com/claraverse-space/ClaraVerse.git
cd ClaraVerse
```

## Step 2: Create Your Environment File

```bash
cp .env.example .env
```

## Step 3: Generate Secrets

Two secrets are required. Generate them and update your `.env` file:

```bash
# Generate JWT signing key
openssl rand -hex 32
# Copy the output and set JWT_SECRET= in .env

# Generate encryption key for user data
openssl rand -hex 32
# Copy the output and set ENCRYPTION_MASTER_KEY= in .env
```

Or do it in one shot:

```bash
sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$(openssl rand -hex 32)|" .env
sed -i "s|^ENCRYPTION_MASTER_KEY=.*|ENCRYPTION_MASTER_KEY=$(openssl rand -hex 32)|" .env
```

:::warning
Keep your `ENCRYPTION_MASTER_KEY` safe. Losing it means losing access to all encrypted user data (conversations, saved credentials). Back it up.
:::

## Step 4: Start ClaraVerse

```bash
docker compose up --build
```

First build takes a few minutes to download images and compile. Subsequent starts are faster.

To run in the background:

```bash
docker compose up --build -d
```

## Step 5: Verify

Once all containers are healthy, open your browser:

| Service | URL |
|---------|-----|
| **ClaraVerse** | [http://localhost](http://localhost) |
| **Backend API** | [http://localhost:3001/health](http://localhost:3001/health) |

You should see the ClaraVerse registration page. The first user to register automatically becomes the admin.

## What Gets Started

Docker Compose brings up six services:

| Container | Image | Port | Purpose |
|-----------|-------|------|---------|
| `claraverse-frontend` | Built from `./frontend` | 80 | React web UI (nginx) |
| `claraverse-backend` | Built from `./backend` | 3001 | Go API server |
| `claraverse-mongodb` | `mongo:7` | 27017 | Conversations, workflows, data |
| `claraverse-mysql` | `mysql:8.0` | 3306 | Providers, models, capabilities |
| `claraverse-redis` | `redis:7-alpine` | 6379 | Job scheduling, WebSocket pub/sub |
| `claraverse-searxng` | `searxng/searxng:latest` | -- | Private web search (internal only) |

## Troubleshooting

**Port 80 already in use?** Another web server (nginx, Apache) may be running. Stop it, or change the frontend port in `docker-compose.yml`:

```yaml
frontend:
  ports:
    - "8080:80"  # Access at http://localhost:8080 instead
```

**Containers failing health checks?** Give them more time on first boot. Check logs:

```bash
docker compose logs -f backend
docker compose logs -f mysql
```

**Out of memory?** The stack uses approximately 4 GB. Close other heavy applications, or increase Docker's memory allocation in Docker Desktop settings.

## Next Steps

Proceed to [Configuration](./configuration.md) to customize environment variables, or jump straight to the [Quickstart](./quickstart.md) to start using ClaraVerse.
