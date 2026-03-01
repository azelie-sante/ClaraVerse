# ClaraVerse API Reference

Complete REST and WebSocket API documentation for ClaraVerseAI.

## Base URLs

| Environment | REST API | WebSocket |
|-------------|----------|-----------|
| Development | `http://localhost:3001` | `ws://localhost:3001` |
| Production | `https://api.yourdomain.com` | `wss://api.yourdomain.com` |

## Authentication

ClaraVerse uses JWT-based authentication. Include the access token in requests:

```
Authorization: Bearer <access_token>
```

For WebSocket connections, pass the token as a query parameter:
```
ws://localhost:3001/ws/chat?token=<access_token>
```

---

## Authentication Endpoints

### Register

Create a new user account.

```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "John Doe"
}
```

**Response:**
```json
{
  "user": {
    "id": "user_abc123",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "tier": "free",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

### Login

Authenticate and receive tokens.

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "user": {
    "id": "user_abc123",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "tier": "pro"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

### Refresh Token

Get new access token using refresh token.

```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

### Logout

Invalidate refresh token.

```http
POST /api/auth/logout
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Get Current User

```http
GET /api/auth/me
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "id": "user_abc123",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "user",
  "tier": "pro",
  "preferences": {},
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

## Agent Endpoints

### List Agents

```http
GET /api/agents
Authorization: Bearer <access_token>
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | int | Results per page (default: 20) |
| `offset` | int | Pagination offset |
| `sort` | string | Sort field (created_at, updated_at, name) |

**Response:**
```json
{
  "agents": [
    {
      "id": "agent_xyz789",
      "name": "Customer Support Agent",
      "description": "Handles customer inquiries",
      "model": "gpt-4o",
      "tools": ["search_web", "send_email"],
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-16T14:20:00Z"
    }
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

### Create Agent

```http
POST /api/agents
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "My Agent",
  "description": "Agent description",
  "model": "gpt-4o",
  "system_prompt": "You are a helpful assistant.",
  "tools": ["search_web", "calculate_math"],
  "temperature": 0.7
}
```

### Get Agent

```http
GET /api/agents/:id
Authorization: Bearer <access_token>
```

### Update Agent

```http
PUT /api/agents/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Updated Agent Name",
  "tools": ["search_web", "send_slack"]
}
```

### Delete Agent

```http
DELETE /api/agents/:id
Authorization: Bearer <access_token>
```

### Sync Agent

Upload a local agent to the backend.

```http
POST /api/agents/:id/sync
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "agent": { ... },
  "workflow": { ... }
}
```

---

## Workflow Endpoints

### Get Workflow

```http
GET /api/agents/:id/workflow
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "id": "workflow_abc123",
  "agent_id": "agent_xyz789",
  "nodes": [
    {
      "id": "node_1",
      "type": "trigger",
      "position": { "x": 100, "y": 100 },
      "data": { "trigger_type": "manual" }
    },
    {
      "id": "node_2",
      "type": "llm",
      "position": { "x": 300, "y": 100 },
      "data": { "model": "gpt-4o", "prompt": "..." }
    }
  ],
  "edges": [
    { "source": "node_1", "target": "node_2" }
  ],
  "version": 3,
  "updated_at": "2024-01-16T14:20:00Z"
}
```

### Save Workflow

```http
PUT /api/agents/:id/workflow
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "nodes": [...],
  "edges": [...]
}
```

### Generate Workflow

AI-assisted workflow generation from description.

```http
POST /api/agents/:id/generate-workflow
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "description": "Create a workflow that searches the web and summarizes results"
}
```

### List Workflow Versions

```http
GET /api/agents/:id/workflow/versions
Authorization: Bearer <access_token>
```

### Restore Workflow Version

```http
POST /api/agents/:id/workflow/restore/:version
Authorization: Bearer <access_token>
```

---

## Schedule Endpoints

### Create Schedule

```http
POST /api/agents/:id/schedule
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "cron": "0 9 * * *",
  "timezone": "America/New_York",
  "enabled": true,
  "input": { "query": "daily news" }
}
```

### Get Schedule

```http
GET /api/agents/:id/schedule
Authorization: Bearer <access_token>
```

### Update Schedule

```http
PUT /api/agents/:id/schedule
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "cron": "0 10 * * MON-FRI",
  "enabled": true
}
```

### Delete Schedule

```http
DELETE /api/agents/:id/schedule
Authorization: Bearer <access_token>
```

### Trigger Now

Manually trigger a scheduled agent.

```http
POST /api/agents/:id/schedule/run
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "input": { "override": "value" }
}
```

---

## Execution Endpoints

### List Executions (by Agent)

```http
GET /api/agents/:id/executions
Authorization: Bearer <access_token>
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | int | Results per page (default: 20) |
| `offset` | int | Pagination offset |
| `status` | string | Filter by status (pending, running, completed, failed) |

### List All Executions

```http
GET /api/executions
Authorization: Bearer <access_token>
```

### Get Execution

```http
GET /api/executions/:id
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "id": "exec_abc123",
  "agent_id": "agent_xyz789",
  "status": "completed",
  "trigger": "schedule",
  "input": { "query": "daily news" },
  "output": { "summary": "..." },
  "duration_ms": 2500,
  "started_at": "2024-01-16T09:00:00Z",
  "completed_at": "2024-01-16T09:00:02Z",
  "node_results": [...]
}
```

### Get Execution Stats

```http
GET /api/agents/:id/executions/stats
Authorization: Bearer <access_token>
```

---

## Memory Endpoints

### List Memories

```http
GET /api/memories
Authorization: Bearer <access_token>
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `archived` | bool | Include archived memories |
| `limit` | int | Results per page |

### Create Memory

```http
POST /api/memories
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "content": "User prefers concise responses",
  "category": "preference",
  "importance": "high"
}
```

### Get Memory

```http
GET /api/memories/:id
Authorization: Bearer <access_token>
```

### Update Memory

```http
PUT /api/memories/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "content": "Updated memory content",
  "importance": "medium"
}
```

### Delete Memory

```http
DELETE /api/memories/:id
Authorization: Bearer <access_token>
```

### Archive Memory

```http
POST /api/memories/:id/archive
Authorization: Bearer <access_token>
```

### Unarchive Memory

```http
POST /api/memories/:id/unarchive
Authorization: Bearer <access_token>
```

### Get Memory Stats

```http
GET /api/memories/stats
Authorization: Bearer <access_token>
```

---

## Tool Endpoints

### List All Tools

```http
GET /api/tools
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "tools": [
    {
      "name": "search_web",
      "display_name": "Search Web",
      "description": "Search the web using SearXNG",
      "category": "data_sources",
      "icon": "Search",
      "parameters": {
        "type": "object",
        "properties": {
          "query": { "type": "string", "description": "Search query" }
        },
        "required": ["query"]
      }
    }
  ]
}
```

### Get Available Tools

Returns tools filtered by user's configured credentials.

```http
GET /api/tools/available
Authorization: Bearer <access_token>
```

### Recommend Tools

AI-assisted tool recommendation based on task description.

```http
POST /api/tools/recommend
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "task": "Send a summary to Slack every morning"
}
```

### Get Tool Registry

Returns full tool registry for workflow builder.

```http
GET /api/tools/registry
Authorization: Bearer <access_token>
```

---

## Credential Endpoints

### List Integrations

```http
GET /api/integrations
```

**Response:**
```json
{
  "integrations": [
    {
      "id": "slack",
      "name": "Slack",
      "description": "Send messages to Slack channels",
      "icon": "slack",
      "auth_type": "api_key",
      "fields": [
        { "name": "webhook_url", "type": "string", "required": true }
      ]
    }
  ]
}
```

### List Credentials

```http
GET /api/credentials
Authorization: Bearer <access_token>
```

### Create Credential

```http
POST /api/credentials
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "integration_id": "slack",
  "name": "My Slack Workspace",
  "credentials": {
    "webhook_url": "https://hooks.slack.com/..."
  }
}
```

### Update Credential

```http
PUT /api/credentials/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Updated Name",
  "credentials": { ... }
}
```

### Delete Credential

```http
DELETE /api/credentials/:id
Authorization: Bearer <access_token>
```

### Test Credential

```http
POST /api/credentials/:id/test
Authorization: Bearer <access_token>
```

---

## Chat Sync Endpoints

Optional server-side chat persistence (encrypted).

### Sync All Chats

```http
GET /api/chats/sync
Authorization: Bearer <access_token>
```

### Bulk Sync

```http
POST /api/chats/sync
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "chats": [...]
}
```

### Create/Update Chat

```http
POST /api/chats
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "id": "chat_local_123",
  "title": "My Chat",
  "messages": [...],
  "model": "gpt-4o"
}
```

### Get Chat

```http
GET /api/chats/:id
Authorization: Bearer <access_token>
```

### Delete Chat

```http
DELETE /api/chats/:id
Authorization: Bearer <access_token>
```

### Delete All Chats

GDPR compliance - delete all user chats.

```http
DELETE /api/chats
Authorization: Bearer <access_token>
```

---

## API Key Endpoints

Manage API keys for external triggers.

### Create API Key

```http
POST /api/keys
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Production Key",
  "scopes": ["trigger", "upload"],
  "expires_at": "2025-01-01T00:00:00Z"
}
```

**Response:**
```json
{
  "id": "key_abc123",
  "name": "Production Key",
  "key": "cv_live_abc123xyz...",
  "scopes": ["trigger", "upload"],
  "created_at": "2024-01-16T10:00:00Z",
  "expires_at": "2025-01-01T00:00:00Z"
}
```

> **Note:** The full key is only shown once at creation.

### List API Keys

```http
GET /api/keys
Authorization: Bearer <access_token>
```

### Revoke API Key

```http
POST /api/keys/:id/revoke
Authorization: Bearer <access_token>
```

### Delete API Key

```http
DELETE /api/keys/:id
Authorization: Bearer <access_token>
```

---

## Trigger Endpoints

External API for triggering agents. Requires API key authentication.

### Trigger Agent

```http
POST /api/trigger/:agentId
X-API-Key: cv_live_abc123xyz...
Content-Type: application/json

{
  "input": {
    "message": "Process this data",
    "data": { ... }
  },
  "webhook_url": "https://your-server.com/webhook"
}
```

**Response:**
```json
{
  "execution_id": "exec_xyz789",
  "status": "pending",
  "message": "Execution started"
}
```

### Get Execution Status

```http
GET /api/trigger/status/:executionId
X-API-Key: cv_live_abc123xyz...
```

---

## Upload Endpoints

### Upload File

```http
POST /api/upload
Authorization: Bearer <access_token>
Content-Type: multipart/form-data

file: <binary>
```

**Response:**
```json
{
  "id": "file_abc123",
  "filename": "document.pdf",
  "size": 102400,
  "mime_type": "application/pdf",
  "url": "/uploads/file_abc123.pdf",
  "expires_at": "2024-01-17T10:00:00Z"
}
```

### Check File Status

```http
GET /api/upload/:id/status
Authorization: Bearer <access_token>
```

### Delete Upload

```http
DELETE /api/upload/:id
Authorization: Bearer <access_token>
```

### External Upload

For API key authenticated uploads.

```http
POST /api/external/upload
X-API-Key: cv_live_abc123xyz...
Content-Type: multipart/form-data

file: <binary>
```

---

## Admin Endpoints

Requires admin role or SUPERADMIN_USER_IDS.

### Get Admin Status

```http
GET /api/admin/me
Authorization: Bearer <access_token>
```

### List Users

```http
GET /api/admin/users
Authorization: Bearer <access_token>
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | int | Results per page |
| `offset` | int | Pagination offset |
| `search` | string | Search by email/name |
| `tier` | string | Filter by tier |

### Get User Details

```http
GET /api/admin/users/:userID
Authorization: Bearer <access_token>
```

### Set Limit Overrides

```http
POST /api/admin/users/:userID/overrides
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "messages_per_day": 1000,
  "executions_per_day": 100,
  "storage_mb": 5000
}
```

### Remove All Overrides

```http
DELETE /api/admin/users/:userID/overrides
Authorization: Bearer <access_token>
```

### Analytics Endpoints

```http
GET /api/admin/analytics/overview
GET /api/admin/analytics/providers
GET /api/admin/analytics/chats
GET /api/admin/analytics/models
GET /api/admin/analytics/agents
Authorization: Bearer <access_token>
```

### Provider Management

```http
GET /api/admin/providers
POST /api/admin/providers
PUT /api/admin/providers/:id
DELETE /api/admin/providers/:id
PUT /api/admin/providers/:id/toggle
Authorization: Bearer <access_token>
```

### Model Management

```http
GET /api/admin/models
POST /api/admin/models
PUT /api/admin/models/:modelId
DELETE /api/admin/models/:modelId
POST /api/admin/models/:modelId/test/connection
POST /api/admin/models/:modelId/benchmark
Authorization: Bearer <access_token>
```

---

## WebSocket Endpoints

### Chat WebSocket

Real-time streaming chat.

```
ws://localhost:3001/ws/chat?token=<access_token>
```

**Client → Server Messages:**

```json
{
  "type": "chat",
  "payload": {
    "conversation_id": "conv_123",
    "message": "Hello, how are you?",
    "model": "gpt-4o",
    "tools": ["search_web"],
    "stream": true
  }
}
```

**Server → Client Messages:**

```json
// Stream chunk
{
  "type": "stream",
  "payload": {
    "conversation_id": "conv_123",
    "content": "Hello! I'm",
    "done": false
  }
}

// Tool call
{
  "type": "tool_call",
  "payload": {
    "conversation_id": "conv_123",
    "tool": "search_web",
    "arguments": { "query": "weather today" },
    "status": "executing"
  }
}

// Tool result
{
  "type": "tool_result",
  "payload": {
    "conversation_id": "conv_123",
    "tool": "search_web",
    "result": "...",
    "status": "completed"
  }
}

// Stream complete
{
  "type": "stream",
  "payload": {
    "conversation_id": "conv_123",
    "content": "",
    "done": true,
    "usage": {
      "prompt_tokens": 150,
      "completion_tokens": 200
    }
  }
}

// Error
{
  "type": "error",
  "payload": {
    "conversation_id": "conv_123",
    "error": "Rate limit exceeded",
    "code": "RATE_LIMIT"
  }
}
```

### Workflow WebSocket

Real-time workflow execution.

```
ws://localhost:3001/ws/workflow?token=<access_token>
```

**Client → Server:**

```json
{
  "type": "execute",
  "payload": {
    "agent_id": "agent_xyz789",
    "input": { "query": "test" }
  }
}
```

**Server → Client:**

```json
// Node started
{
  "type": "node_start",
  "payload": {
    "execution_id": "exec_123",
    "node_id": "node_1",
    "node_type": "llm"
  }
}

// Node completed
{
  "type": "node_complete",
  "payload": {
    "execution_id": "exec_123",
    "node_id": "node_1",
    "output": { ... },
    "duration_ms": 1200
  }
}

// Execution complete
{
  "type": "execution_complete",
  "payload": {
    "execution_id": "exec_123",
    "status": "completed",
    "output": { ... },
    "duration_ms": 3500
  }
}
```

### MCP WebSocket

Model Context Protocol bridge.

```
ws://localhost:3001/mcp/connect?token=<access_token>
```

**Client → Server:**

```json
{
  "jsonrpc": "2.0",
  "method": "tools/list",
  "id": 1
}
```

**Server → Client:**

```json
{
  "jsonrpc": "2.0",
  "result": {
    "tools": [...]
  },
  "id": 1
}
```

---

## Public Endpoints

These endpoints don't require authentication:

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Health check |
| `GET /api/providers` | List providers |
| `GET /api/models` | List models |
| `GET /api/integrations` | List integrations |
| `GET /api/config/recommended-models` | Recommended model config |
| `GET /api/privacy-policy` | Privacy policy |
| `GET /api/proxy/image` | Image proxy (rate limited) |

---

## Error Responses

All errors follow this format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": { ... }
}
```

**Common Error Codes:**

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `RATE_LIMIT` | 429 | Too many requests |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Rate Limits

| Endpoint Type | Default Limit | Config Variable |
|---------------|---------------|-----------------|
| Global API | 200/min | `RATE_LIMIT_GLOBAL_API` |
| Public Read | 120/min | `RATE_LIMIT_PUBLIC_READ` |
| Authenticated | 60/min | `RATE_LIMIT_AUTHENTICATED` |
| WebSocket | 20/min | `RATE_LIMIT_WEBSOCKET` |
| Image Proxy | 60/min | `RATE_LIMIT_IMAGE_PROXY` |
| Upload | 10/min | Hardcoded |

---

## Related Documentation

- [Architecture Guide](ARCHITECTURE.md) - System design
- [Developer Guide](DEVELOPER_GUIDE.md) - Local setup
- [Security Guide](FINAL_SECURITY_INSPECTION.md) - Security details
- [Admin Guide](ADMIN_GUIDE.md) - Administration
- [Quick Reference](QUICK_REFERENCE.md) - Common commands
