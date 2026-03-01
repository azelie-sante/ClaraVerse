# ClaraVerse Admin Guide

Complete guide for system administrators managing ClaraVerseAI.

## Becoming an Admin

### Method 1: Role Field (Recommended)

Set the user's role to `admin` in MongoDB:

```javascript
// MongoDB shell
db.users.updateOne(
  { email: "admin@example.com" },
  { $set: { role: "admin" } }
)
```

### Method 2: Environment Variable (Legacy)

Add user IDs to `SUPERADMIN_USER_IDS`:

```bash
# .env
SUPERADMIN_USER_IDS=user-id-1,user-id-2,user-id-3
```

### Default Admin Account

On first startup, a default admin is created:

```
Email: admin@localhost
Password: admin
```

**Change this password immediately after first login.**

---

## Admin Panel Access

Access the admin panel at:
- **Development:** http://localhost:5173/admin
- **Production:** https://yourdomain.com/admin

### Verify Admin Status

```http
GET /api/admin/me
Authorization: Bearer <access_token>
```

Response:
```json
{
  "is_admin": true,
  "user_id": "user_abc123",
  "email": "admin@example.com"
}
```

---

## User Management

### List Users

```http
GET /api/admin/users?limit=20&offset=0&search=john
Authorization: Bearer <access_token>
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | int | Results per page (default: 20, max: 100) |
| `offset` | int | Pagination offset |
| `search` | string | Search by email or name |
| `tier` | string | Filter by tier (free, pro, max, enterprise) |
| `sort` | string | Sort field (created_at, email, tier) |

**Response:**
```json
{
  "users": [
    {
      "id": "user_abc123",
      "email": "john@example.com",
      "name": "John Doe",
      "tier": "pro",
      "role": "user",
      "created_at": "2024-01-15T10:30:00Z",
      "last_login": "2024-01-20T14:00:00Z",
      "stats": {
        "agents_count": 5,
        "executions_count": 150,
        "storage_used_mb": 250
      }
    }
  ],
  "total": 150,
  "limit": 20,
  "offset": 0
}
```

### Get User Details

```http
GET /api/admin/users/:userID
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "user": {
    "id": "user_abc123",
    "email": "john@example.com",
    "name": "John Doe",
    "tier": "pro",
    "role": "user",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "limits": {
    "messages_per_day": 1000,
    "executions_per_day": 100,
    "storage_mb": 5000,
    "retention_days": 365
  },
  "overrides": {
    "messages_per_day": 2000
  },
  "usage": {
    "messages_today": 450,
    "executions_today": 23,
    "storage_used_mb": 250
  }
}
```

### Set User Limit Overrides

Override tier limits for specific users:

```http
POST /api/admin/users/:userID/overrides
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "messages_per_day": 2000,
  "executions_per_day": 200,
  "storage_mb": 10000
}
```

### Remove All Overrides

```http
DELETE /api/admin/users/:userID/overrides
Authorization: Bearer <access_token>
```

---

## Provider Management

### List Providers

```http
GET /api/admin/providers
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "providers": [
    {
      "id": 1,
      "name": "OpenAI",
      "type": "openai",
      "base_url": "https://api.openai.com/v1",
      "api_key_configured": true,
      "enabled": true,
      "models_count": 15
    },
    {
      "id": 2,
      "name": "Anthropic",
      "type": "anthropic",
      "base_url": "https://api.anthropic.com",
      "api_key_configured": true,
      "enabled": true,
      "models_count": 5
    }
  ]
}
```

### Create Provider

```http
POST /api/admin/providers
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Local Ollama",
  "type": "openai",
  "base_url": "http://localhost:11434/v1",
  "api_key": "",
  "enabled": true
}
```

### Update Provider

```http
PUT /api/admin/providers/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Updated Name",
  "base_url": "https://new-url.com/v1",
  "enabled": true
}
```

### Toggle Provider

Enable/disable a provider:

```http
PUT /api/admin/providers/:id/toggle
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "enabled": false
}
```

### Delete Provider

```http
DELETE /api/admin/providers/:id
Authorization: Bearer <access_token>
```

### Fetch Models from Provider

Discover available models from a provider's API:

```http
POST /api/admin/providers/:providerId/fetch
Authorization: Bearer <access_token>
```

---

## Model Management

### List All Models

```http
GET /api/admin/models
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "models": [
    {
      "id": 1,
      "provider_id": 1,
      "model_id": "gpt-4o",
      "display_name": "GPT-4o",
      "context_length": 128000,
      "supports_tools": true,
      "supports_vision": true,
      "supports_streaming": true,
      "visible": true,
      "agents_enabled": true,
      "tier": null
    }
  ]
}
```

### Create Model

```http
POST /api/admin/models
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "provider_id": 1,
  "model_id": "gpt-4o-mini",
  "display_name": "GPT-4o Mini",
  "context_length": 128000,
  "supports_tools": true,
  "supports_vision": true,
  "supports_streaming": true,
  "visible": true,
  "agents_enabled": true
}
```

### Update Model

```http
PUT /api/admin/models/:modelId
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "display_name": "Updated Name",
  "visible": true,
  "agents_enabled": false
}
```

### Bulk Update Models

Enable/disable agents for multiple models:

```http
PUT /api/admin/models/bulk/agents-enabled
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "model_ids": [1, 2, 3],
  "agents_enabled": true
}
```

Update visibility:

```http
PUT /api/admin/models/bulk/visibility
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "model_ids": [1, 2, 3],
  "visible": false
}
```


### Test Model Connection

```http
POST /api/admin/models/:modelId/test/connection
Authorization: Bearer <access_token>
```

### Run Model Benchmark

```http
POST /api/admin/models/:modelId/benchmark
Authorization: Bearer <access_token>
```

---

## Model Aliases

Aliases allow alternative names for models.

### List Aliases

```http
GET /api/admin/models/:modelId/aliases
Authorization: Bearer <access_token>
```

### Create Alias

```http
POST /api/admin/models/:modelId/aliases
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "alias": "gpt4",
  "description": "Shorthand for GPT-4o"
}
```

### Delete Alias

```http
DELETE /api/admin/models/:modelId/aliases/:alias
Authorization: Bearer <access_token>
```

### Import Aliases from JSON

```http
POST /api/admin/models/import-aliases
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "aliases": [
    { "model_id": "gpt-4o", "alias": "gpt4" },
    { "model_id": "claude-3-5-sonnet-20241022", "alias": "sonnet" }
  ]
}
```

---

## Analytics

### Overview Analytics

```http
GET /api/admin/analytics/overview
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "total_users": 1500,
  "active_users_today": 350,
  "active_users_week": 800,
  "total_agents": 2500,
  "total_executions": 150000,
  "executions_today": 2500,
  "storage_used_gb": 125.5
}
```

### Provider Analytics

```http
GET /api/admin/analytics/providers
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "providers": [
    {
      "provider_id": 1,
      "provider_name": "OpenAI",
      "total_requests": 50000,
      "requests_today": 1500,
      "tokens_used": 25000000,
      "avg_response_time_ms": 850
    }
  ]
}
```

### Chat Analytics

```http
GET /api/admin/analytics/chats
Authorization: Bearer <access_token>
```

### Model Analytics

```http
GET /api/admin/analytics/models
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "models": [
    {
      "model_id": "gpt-4o",
      "total_requests": 25000,
      "requests_today": 800,
      "avg_tokens_per_request": 500,
      "success_rate": 99.5
    }
  ]
}
```

### Agent Analytics

```http
GET /api/admin/analytics/agents
Authorization: Bearer <access_token>
```

---

## System Operations

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "2.0.0",
  "uptime": "72h15m30s",
  "services": {
    "mysql": "connected",
    "mongodb": "connected",
    "redis": "connected"
  }
}
```

### System Stats (Legacy)

```http
GET /api/admin/stats
Authorization: Bearer <access_token>
```

---

## Background Jobs

ClaraVerse runs scheduled maintenance jobs:

| Job | Schedule | Description |
|-----|----------|-------------|
| Retention Cleanup | Daily 2 AM UTC | Delete expired data per tier retention |
| Grace Period Check | Hourly | Handle tier downgrades |
| Promo Expiration | Hourly | Expire promotional tiers |

### Monitor Jobs

Check backend logs for job execution:
```
✅ Background job scheduler started
🕐 Background jobs: retention cleanup (daily 2 AM), grace period check (hourly)
```

---

## Database Operations

### MongoDB Collections

```javascript
// Users
db.users.find({ tier: "pro" }).count()

// Agents
db.agents.find({ user_id: "user_abc123" })

// Executions
db.executions.find({ 
  created_at: { $gte: ISODate("2024-01-01") } 
}).count()

// Clean up old executions
db.executions.deleteMany({ 
  created_at: { $lt: ISODate("2024-01-01") } 
})
```

### MySQL Tables

```sql
-- List providers
SELECT * FROM providers;

-- List models with provider
SELECT m.*, p.name as provider_name 
FROM models m 
JOIN providers p ON m.provider_id = p.id;

-- Model usage stats
SELECT model_id, COUNT(*) as usage_count 
FROM model_usage 
GROUP BY model_id 
ORDER BY usage_count DESC;
```

### Redis Keys

```bash
# List scheduler keys
redis-cli KEYS "scheduler:*"

# Check pubsub channels
redis-cli PUBSUB CHANNELS

# Clear rate limit for user
redis-cli DEL "ratelimit:user_abc123"
```

---

## Troubleshooting

### Common Issues

**Users can't login:**
1. Check JWT_SECRET is set
2. Verify MongoDB connection
3. Check user exists: `db.users.findOne({ email: "..." })`

**Models not appearing:**
1. Check provider is enabled
2. Verify model visibility: `visible: true`
3. Check tier restrictions

**Rate limit errors:**
1. Check RATE_LIMIT_* env vars
2. Clear Redis rate limit keys
3. Adjust limits if needed

**WebSocket disconnections:**
1. Check RATE_LIMIT_WEBSOCKET
2. Verify CORS settings
3. Check Redis pub/sub connection

### Logs

```bash
# Docker logs
docker logs claraverse-backend -f

# Filter for errors
docker logs claraverse-backend 2>&1 | grep "ERROR\|FATAL"

# Filter for security events
docker logs claraverse-backend 2>&1 | grep "🚫\|⚠️"
```

---

## Security Considerations

### Admin Best Practices

1. **Use strong passwords** for admin accounts
2. **Limit SUPERADMIN_USER_IDS** to necessary users
3. **Audit admin actions** via logs
4. **Rotate secrets** periodically
5. **Monitor analytics** for anomalies

### Sensitive Operations

These operations should be used carefully:

| Operation | Risk | Mitigation |
|-----------|------|------------|
| Delete provider | Breaks existing agents | Disable first, migrate |
| Bulk visibility change | Hides models from users | Communicate changes |
| User limit overrides | Can increase costs | Document reasons |
| Database direct access | Data corruption | Backup first |

---

## Related Documentation

- [Architecture Guide](ARCHITECTURE.md) - System design
- [API Reference](API_REFERENCE.md) - API documentation
- [Developer Guide](DEVELOPER_GUIDE.md) - Local setup
- [Security Guide](FINAL_SECURITY_INSPECTION.md) - Security details
- [Quick Reference](QUICK_REFERENCE.md) - Common commands
