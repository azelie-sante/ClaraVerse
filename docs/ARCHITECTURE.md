# ClaraVerse Architecture Guide

This document describes the system architecture, components, and data flow of ClaraVerseAI.

## Overview

ClaraVerse is a private AI workspace built as a monorepo with a Go backend and React frontend. It provides chat, image generation, and visual workflow automation while maintaining user privacy through browser-local storage.

```
ClaraVerseAI/
├── backend/           # Go 1.24+ API server
├── frontend/          # React 19 + Vite + TypeScript
├── searxng/           # Self-hosted search configuration
└── docker-compose.yml # Service orchestration
```

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend                                 │
│  React 19 + TypeScript + Vite + Tailwind CSS 4 + Zustand        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  IndexedDB  │  │  WebSocket  │  │  REST API   │             │
│  │ (Chat Data) │  │   Client    │  │   Client    │             │
│  └─────────────┘  └──────┬──────┘  └──────┬──────┘             │
└──────────────────────────┼────────────────┼─────────────────────┘
                           │                │
                           ▼                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Backend (Go Fiber)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  WebSocket  │  │ REST API    │  │ Middleware  │             │
│  │  Handlers   │  │ Handlers    │  │ (Auth/Rate) │             │
│  └──────┬──────┘  └──────┬──────┘  └─────────────┘             │
│         │                │                                       │
│  ┌──────┴────────────────┴──────┐                               │
│  │         Services Layer        │                               │
│  │  Chat │ Agent │ Workflow │ Mem│                               │
│  └──────────────┬───────────────┘                               │
│                 │                                                │
│  ┌──────────────┴───────────────┐                               │
│  │       Tool Registry (80+)     │                               │
│  │  Built-in │ MCP │ Composio   │                               │
│  └──────────────────────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
         │              │              │              │
         ▼              ▼              ▼              ▼
    ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
    │  MySQL  │   │ MongoDB │   │  Redis  │   │ SearXNG │
    │Provider │   │  Users  │   │Scheduler│   │ Search  │
    │ Models  │   │ Agents  │   │ PubSub  │   │         │
    └─────────┘   └─────────┘   └─────────┘   └─────────┘
                                                    │
                                              ┌─────────┐
                                              │   E2B   │
                                              │  Code   │
                                              │  Exec   │
                                              └─────────┘
```

## Components

### Frontend (React 19)

| Component | Description |
|-----------|-------------|
| **Chat** | Real-time streaming chat with IndexedDB local storage |
| **Agent Builder** | Visual workflow editor for creating AI agents |
| **Admin Panel** | User management, analytics, provider configuration |
| **Credentials** | Secure integration credential management |
| **Memory** | View and manage Clara's memory system |
| **Settings** | User preferences and API key configuration |

**Key Technologies:**
- React 19 with TypeScript
- Vite for bundling
- Tailwind CSS 4 for styling
- Zustand for state management
- IndexedDB for zero-knowledge chat storage

### Backend (Go Fiber)

| Layer | Components |
|-------|------------|
| **Handlers** | HTTP/WebSocket request handlers |
| **Middleware** | Auth, rate limiting, CORS, admin checks |
| **Services** | Business logic (Chat, Agent, Workflow, Memory, etc.) |
| **Models** | Data structures and database schemas |
| **Tools** | 80+ built-in tool implementations |

**Key Technologies:**
- Go 1.24+
- Fiber (Express-inspired web framework)
- WebSocket for real-time streaming
- JWT for authentication

### Databases

#### MySQL (Relational Data)
Stores provider configurations and model metadata.

```sql
-- Core tables
providers        -- AI provider configurations (OpenAI, Anthropic, etc.)
models           -- Available models per provider
model_aliases    -- Alternative names for models
model_capabilities -- Per-model feature flags

-- MCP tables
mcp_servers      -- MCP server configurations
mcp_server_tools -- Tools exposed by MCP servers
```

#### MongoDB (Document Data)
Stores user data, agents, and workflows.

```javascript
// Collections
users            // User accounts and preferences
agents           // Agent definitions and configurations
workflows        // Visual workflow definitions
workflow_versions // Workflow version history
executions       // Agent execution history
credentials      // Encrypted integration credentials
memories         // Clara's memory system
chat_sessions    // Optional chat sync (encrypted)
analytics        // Usage analytics
```

#### Redis (Ephemeral Data)
Handles job scheduling and real-time messaging.

```
scheduler:*      // Cron job schedules
pubsub:*         // WebSocket message routing
session:*        // Session data
```

### External Services

| Service | Purpose |
|---------|---------|
| **SearXNG** | Self-hosted meta-search engine for web searches |
| **E2B** | Sandboxed code execution (local Docker or cloud) |
| **Composio** | OAuth integrations (Google Sheets, Gmail) |

## Data Flow

### Chat Message Flow

```
1. User sends message (Frontend)
   └─▶ IndexedDB stores locally
   └─▶ WebSocket sends to backend

2. Backend receives message
   └─▶ JWT validation
   └─▶ Rate limit check
   └─▶ Chat service processes

3. LLM streaming response
   └─▶ Tool calls detected → Tool execution
   └─▶ Stream chunks via WebSocket

4. Frontend receives stream
   └─▶ Real-time UI update
   └─▶ IndexedDB persistence
```

### Workflow Execution Flow

```
1. Trigger (Manual/Schedule/API)
   └─▶ Load agent + workflow definition

2. Node execution (sequential/parallel)
   └─▶ Input transformation
   └─▶ Tool/LLM execution
   └─▶ Output mapping

3. Results
   └─▶ Execution history saved
   └─▶ Webhook notifications (if configured)
```

## Authentication Architecture

ClaraVerse v2.0 uses local JWT authentication (moved away from Supabase).

```
┌──────────────────────────────────────────────────────────────┐
│                    Authentication Flow                        │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Register                                                     │
│  ┌────────┐     ┌──────────┐     ┌──────────┐              │
│  │ Client │────▶│ /register│────▶│ Argon2id │              │
│  └────────┘     └──────────┘     │  Hash    │              │
│                                   └────┬─────┘              │
│                                        ▼                     │
│                                   ┌──────────┐              │
│                                   │ MongoDB  │              │
│                                   │  users   │              │
│                                   └──────────┘              │
│                                                              │
│  Login                                                        │
│  ┌────────┐     ┌──────────┐     ┌──────────┐              │
│  │ Client │────▶│  /login  │────▶│ Verify   │              │
│  └────────┘     └──────────┘     │ Password │              │
│       ▲                          └────┬─────┘              │
│       │                               ▼                     │
│       │         ┌──────────────────────────┐               │
│       └─────────│ JWT Access + Refresh     │               │
│                 │ Tokens (HS256)           │               │
│                 └──────────────────────────┘               │
│                                                              │
│  Authenticated Request                                       │
│  ┌────────┐     ┌──────────┐     ┌──────────┐              │
│  │ Client │────▶│Middleware│────▶│  Verify  │              │
│  │ + JWT  │     │          │     │   JWT    │              │
│  └────────┘     └──────────┘     └────┬─────┘              │
│                                        ▼                     │
│                                   ┌──────────┐              │
│                                   │ Handler  │              │
│                                   └──────────┘              │
└──────────────────────────────────────────────────────────────┘
```

**Token Configuration:**
- Access Token: 15 minutes (configurable via `JWT_ACCESS_TOKEN_EXPIRY`)
- Refresh Token: 7 days (configurable via `JWT_REFRESH_TOKEN_EXPIRY`)

## Encryption Architecture

User data is encrypted with AES-256-GCM using per-user derived keys.

```
┌─────────────────────────────────────────────────────────────┐
│                   Encryption Flow                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ENCRYPTION_MASTER_KEY (env)                                │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │      HKDF       │◀── User ID (salt)                     │
│  │ Key Derivation  │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐     ┌─────────────────┐              │
│  │ Per-User Key    │────▶│   AES-256-GCM   │              │
│  └─────────────────┘     │   Encryption    │              │
│                          └────────┬────────┘              │
│                                   │                        │
│                                   ▼                        │
│                          ┌─────────────────┐              │
│                          │ Encrypted Data  │              │
│                          │ (credentials,   │              │
│                          │  chat sync)     │              │
│                          └─────────────────┘              │
└─────────────────────────────────────────────────────────────┘
```

## Tool Architecture

ClaraVerse includes 80+ built-in tools organized by category.

### Tool Categories

| Category | Tools |
|----------|-------|
| **Search & Data** | Web search, image search, web scraper |
| **Computation** | Math, Python runner, data analyst, ML trainer |
| **Time** | Current time, timezone conversion |
| **Documents** | PDF generation, document reading, spreadsheets |
| **Communication** | Slack, Discord, Telegram, Teams, Email |
| **CRM** | HubSpot, LeadSquared |
| **Project Management** | ClickUp, Jira, Linear, Trello, Asana |
| **Development** | GitHub, GitLab, API testing |
| **E-commerce** | Shopify products/orders/customers |
| **Analytics** | Mixpanel, PostHog |
| **Storage** | AWS S3, MongoDB, Redis |
| **Google Workspace** | Sheets, Gmail (via Composio) |

### Tool Registration

```go
// Built-in tools are registered at startup
func registerBuiltInTools(r *Registry) {
    r.Register(NewSearchTool())
    r.Register(NewMathTool())
    // ... 80+ tools
}

// User-specific MCP tools
r.RegisterUserTool(userID, mcpTool)
```

### MCP Bridge

The Model Context Protocol (MCP) bridge allows connecting external tools.

```
┌────────────┐     WebSocket      ┌────────────┐
│  Frontend  │◀──────────────────▶│  Backend   │
│ MCP Client │                    │ MCP Bridge │
└────────────┘                    └─────┬──────┘
                                        │
                                        ▼
                                  ┌────────────┐
                                  │ MCP Server │
                                  │  (stdio)   │
                                  └────────────┘
```

## Deployment Architecture

### Docker Compose (Recommended)

```yaml
services:
  frontend    # React app (port 80)
  backend     # Go API (port 3001)
  mongodb     # Document store (port 27017)
  mysql       # Relational store (port 3306)
  redis       # Cache/scheduler (port 6379)
  searxng     # Search engine (internal)
  e2b-service # Code execution (internal)
```

### Scaling Considerations

- **Backend**: Horizontally scalable with Redis pub/sub for WebSocket message routing
- **Frontend**: Static assets, CDN-friendly
- **Databases**: Standard MySQL/MongoDB scaling patterns
- **E2B**: Pool size configurable via `E2B_SANDBOX_POOL_SIZE`

## Security Boundaries

```
┌─────────────────────────────────────────────────────────────┐
│                    Public Internet                           │
└─────────────────────────┬───────────────────────────────────┘
                          │
                    ┌─────┴─────┐
                    │   CORS    │
                    │   Check   │
                    └─────┬─────┘
                          │
                    ┌─────┴─────┐
                    │   Rate    │
                    │  Limiter  │
                    └─────┬─────┘
                          │
              ┌───────────┴───────────┐
              │                       │
        ┌─────┴─────┐           ┌─────┴─────┐
        │  Public   │           │ Protected │
        │ Endpoints │           │ Endpoints │
        └───────────┘           └─────┬─────┘
                                      │
                                ┌─────┴─────┐
                                │   JWT     │
                                │   Auth    │
                                └─────┬─────┘
                                      │
                          ┌───────────┴───────────┐
                          │                       │
                    ┌─────┴─────┐           ┌─────┴─────┐
                    │   User    │           │   Admin   │
                    │ Endpoints │           │ Endpoints │
                    └───────────┘           └─────┬─────┘
                                                  │
                                            ┌─────┴─────┐
                                            │   Admin   │
                                            │   Check   │
                                            └───────────┘
```

## Related Documentation

- [API Reference](API_REFERENCE.md) - REST and WebSocket API details
- [Developer Guide](DEVELOPER_GUIDE.md) - Local development setup
- [Security Guide](FINAL_SECURITY_INSPECTION.md) - Security implementation details
- [Admin Guide](ADMIN_GUIDE.md) - System administration
- [Quick Reference](QUICK_REFERENCE.md) - Common commands
