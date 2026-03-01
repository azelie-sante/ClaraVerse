---
title: FAQ
sidebar_label: FAQ
sidebar_position: 99
---

# Frequently Asked Questions

## What AI providers does ClaraVerse support?

ClaraVerse supports **OpenAI**, **Anthropic**, **Google AI (Gemini)**, **Ollama**, and any **OpenAI-compatible endpoint** (LM Studio, vLLM, Groq, Together AI, etc.). You can connect multiple providers at the same time and switch between them freely.

## Can I use ClaraVerse without an API key?

Yes. Connect a local provider like **Ollama** or **LM Studio** and run open-source models entirely on your own hardware. No API key or internet connection required.

## What are the system requirements?

| Requirement | Minimum |
|-------------|---------|
| RAM | 4 GB free |
| Disk | 10 GB free |
| Docker | 20.10+ |
| Docker Compose | v2+ |

ClaraVerse runs six containers (frontend, backend, MongoDB, MySQL, Redis, SearXNG), so make sure Docker has enough memory allocated.

## Where are my conversations stored?

- **Client-side (primary):** The browser uses **IndexedDB** as the primary storage for conversations. All chat data is stored locally in the browser by default.
- **Server-side (opt-in):** If you enable chat sync, MongoDB stores encrypted chat sync data. This is not enabled by default -- the user must opt in.

All data stays on your infrastructure. Nothing is sent to external services.

## How do I back up my data?

Back up the Docker volumes that hold your databases:

```bash
# Stop the stack
docker compose down

# Back up MongoDB data
docker run --rm -v claraverse_mongodb-data:/data -v $(pwd)/backups:/backup \
  alpine tar czf /backup/mongodb-backup.tar.gz -C /data .

# Back up MySQL data
docker run --rm -v claraverse_mysql-data-new:/data -v $(pwd)/backups:/backup \
  alpine tar czf /backup/mysql-backup.tar.gz -C /data .

# Restart
docker compose up -d
```

:::note
The volume name prefix depends on the directory name where you cloned the project (e.g., `claraverse_mongodb-data` if the directory is named `ClaraVerse`). Use `docker volume ls` to verify the exact volume names on your system.
:::

:::tip
Automate backups with a cron job. Also back up your `.env` file -- it contains your `ENCRYPTION_MASTER_KEY`, which is required to decrypt user data.
:::

## How do I add more users?

Users self-register through the ClaraVerse login page. The **first user** to register becomes the admin automatically. All subsequent users get the standard `user` role. Admins can promote users or adjust their limits from the [Admin Panel](./admin/overview.md).

## Is my data private?

Yes. ClaraVerse is fully self-hosted. Your conversations, files, and credentials stay on your own infrastructure. Web search goes through a self-hosted SearXNG instance. The only external calls are to the AI providers you explicitly configure.

:::warning
If you use cloud AI providers (OpenAI, Anthropic, Google), your prompts and responses pass through their APIs. Use local models via Ollama for fully offline, fully private operation.
:::

## How do I update ClaraVerse?

```bash
git pull origin main
docker compose up --build -d
```

This pulls the latest code, rebuilds the containers, and restarts the stack. Your data volumes are preserved across updates.

## Can I use a reverse proxy?

Yes. ClaraVerse works behind **Nginx**, **Caddy**, **Traefik**, and other reverse proxies. Point the proxy at port `80` (frontend) and ensure WebSocket connections to port `3001` (backend) are forwarded correctly.

A minimal Nginx example:

```nginx
server {
    listen 443 ssl;
    server_name clara.example.com;

    location / {
        proxy_pass http://localhost:80;
    }

    location /api/ {
        proxy_pass http://localhost:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## What ports does ClaraVerse use?

| Port | Service |
|------|---------|
| 80 | Frontend (Nginx) |
| 3001 | Backend API + WebSocket |
| 27017 | MongoDB |
| 3306 | MySQL |
| 6379 | Redis |

SearXNG runs on an internal Docker network and is not exposed to the host by default.

:::tip
In production, only expose ports **80** (or 443 via reverse proxy) and **3001** to the network. Keep database ports (27017, 3306, 6379) restricted to localhost or the Docker network.
:::
