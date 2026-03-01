---
title: System Architecture
sidebar_label: Architecture
sidebar_position: 1
---

# System Architecture

ClaraVerse is a self-hosted AI workspace with six core services orchestrated by Docker Compose. All services run on a single bridge network (`claraverse-network`) and communicate internally by container name.

## Service Overview

| Service | Technology | Port | Role |
|---------|-----------|------|------|
| **Frontend** | React 19 / Vite / Nginx | 80 | Web UI, static asset serving |
| **Backend** | Go 1.24+ / Fiber | 3001 | REST API, WebSocket, tool execution |
| **MongoDB** | MongoDB 7 | 27017 | Users, agents, workflows, credentials, memories |
| **MySQL** | MySQL 8.0 | 3306 | AI providers, models, capabilities, MCP servers |
| **Redis** | Redis 7 Alpine | 6379 | Job scheduler, WebSocket pub/sub, sessions |
| **SearXNG** | SearXNG (latest) | 8080 (internal) | Self-hosted meta-search engine |

## Network Diagram

```
                        Internet
                           |
                      Port 80  Port 3001
                        |          |
   ┌────────────────────┼──────────┼────────────────────────┐
   │             claraverse-network (bridge)                 │
   │                    |          |                          │
   │  ┌─────────────┐  │  ┌──────────────┐                  │
   │  │  Frontend    │  │  │   Backend    │                  │
   │  │  (Nginx)     │──┘  │  (Go Fiber)  │                  │
   │  └─────────────┘      └──────┬───────┘                  │
   │                               │                          │
   │          ┌────────┬───────────┼───────────┐              │
   │          ▼        ▼           ▼           ▼              │
   │    ┌─────────┐ ┌───────┐ ┌─────────┐ ┌─────────┐       │
   │    │ MongoDB │ │ MySQL │ │  Redis  │ │ SearXNG │       │
   │    │ :27017  │ │ :3306 │ │ :6379   │ │ :8080   │       │
   │    └─────────┘ └───────┘ └─────────┘ └─────────┘       │
   └──────────────────────────────────────────────────────────┘
```

## Data Flow

### Chat Messages

1. User sends a message in the browser.
2. The frontend stores the message locally in IndexedDB (zero-knowledge design).
3. A WebSocket frame is sent to the backend on port 3001.
4. The backend validates the JWT, checks rate limits, and forwards the message to the configured LLM provider.
5. Streamed response chunks flow back over the same WebSocket connection.
6. If the LLM triggers a tool call (web search, code execution, etc.), the backend executes it and returns the result inline.

### Workflow Execution

1. A trigger fires (manual, scheduled via Redis cron, or API call).
2. The backend loads the agent and workflow definition from MongoDB.
3. Nodes execute sequentially or in parallel, each running tool calls or LLM requests.
4. Execution history is saved to MongoDB. Optional webhook notifications are dispatched.

## What Each Database Stores

### MongoDB (Document Data)

| Collection | Contents |
|------------|----------|
| `users` | Accounts and preferences |
| `agents` | Agent definitions and configurations |
| `workflows` | Visual workflow definitions |
| `workflow_versions` | Version history for workflows |
| `executions` | Agent execution logs |
| `credentials` | Encrypted integration credentials |
| `memories` | Clara memory system entries |
| `chat_sessions` | Optional encrypted chat sync |
| `analytics` | Usage analytics |

### MySQL (Relational Data)

| Table | Contents |
|-------|----------|
| `providers` | AI provider configs (OpenAI, Anthropic, etc.) |
| `models` | Available models per provider |
| `model_aliases` | Alternate names for models |
| `model_capabilities` | Per-model feature flags |
| `mcp_servers` | MCP server configurations |
| `mcp_server_tools` | Tools exposed by MCP servers |

### Redis (Ephemeral Data)

| Key Pattern | Purpose |
|-------------|---------|
| `scheduler:*` | Cron job schedules |
| `pubsub:*` | WebSocket message routing across instances |
| `session:*` | Session data |
| `ratelimit:*` | Rate limiter counters |

## External / Optional Services

| Service | Purpose | Required? |
|---------|---------|-----------|
| **SearXNG** | Web search tool for agents | Yes (included in Compose) |
| **E2B** | Sandboxed Python code execution | No -- requires `E2B_API_KEY` |
| **Composio** | OAuth integrations (Google Sheets, Gmail, etc.) | No -- requires `COMPOSIO_API_KEY` |

## Scaling Notes

- **Backend** is horizontally scalable. Redis pub/sub handles WebSocket message routing across multiple instances.
- **Frontend** serves static assets through Nginx and is CDN-friendly.
- **Databases** follow standard MySQL/MongoDB replication and sharding patterns.
- For a single-server deployment (the typical self-hosting scenario), the default resource limits in `docker-compose.yml` work well for up to ~50 concurrent users.
