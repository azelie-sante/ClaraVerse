---
title: Model Management
sidebar_label: Model Management
sidebar_position: 4
---

# Model Management

Configure which AI models are available to users from **Admin Panel > Models**.

## Listing Models

View all registered models with their provider, capabilities, and visibility status:

```http
GET /api/admin/models
Authorization: Bearer <access_token>
```

Each model entry includes:

| Field | Description |
|-------|-------------|
| `model_id` | The provider's model identifier (e.g., `gpt-4o`) |
| `display_name` | User-facing name shown in model selectors |
| `context_length` | Maximum token context window |
| `supports_tools` | Whether the model supports function/tool calling |
| `supports_vision` | Whether the model accepts image inputs |
| `supports_streaming` | Whether the model supports streamed responses |
| `visible` | Whether users can see and select this model |
| `agents_enabled` | Whether this model can be used with Nexus agents |

## Creating a Model

Register a model manually when auto-fetch does not pick it up, or when you need custom settings:

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

## Updating a Model

Change display name, visibility, capabilities, or agent access:

```http
PUT /api/admin/models/by-id?model_id=<modelId>
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "display_name": "GPT-4o (Latest)",
  "visible": true,
  "agents_enabled": false
}
```

## Bulk Operations

Update multiple models at once.

**Toggle agent access for several models:**

```http
PUT /api/admin/models/bulk/agents-enabled
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "model_ids": [1, 2, 3],
  "agents_enabled": true
}
```

**Toggle visibility for several models:**

```http
PUT /api/admin/models/bulk/visibility
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "model_ids": [1, 2, 3],
  "visible": false
}
```

:::warning
Hiding models affects all users immediately. Communicate changes before bulk-disabling popular models.
:::

## Model Aliases

Aliases give models user-friendly shorthand names. For example, alias `sonnet` for `claude-3-5-sonnet-20241022`.

**List aliases for a model:**

```http
GET /api/admin/models/by-id/aliases?model_id=<modelId>
```

**Create an alias:**

```http
POST /api/admin/models/by-id/aliases?model_id=<modelId>
Content-Type: application/json

{
  "alias": "gpt4",
  "description": "Shorthand for GPT-4o"
}
```

**Delete an alias:**

```http
DELETE /api/admin/models/by-id/aliases?model_id=<modelId>&alias=<alias>
```

**Import aliases in bulk from JSON:**

```http
POST /api/admin/models/import-aliases
Content-Type: application/json

{
  "aliases": [
    { "model_id": "gpt-4o", "alias": "gpt4" },
    { "model_id": "claude-3-5-sonnet-20241022", "alias": "sonnet" }
  ]
}
```

## Testing and Benchmarking

**Test a model's connection** to verify the provider is reachable and the model responds:

```http
POST /api/admin/models/by-id/test/connection?model_id=<modelId>
Authorization: Bearer <access_token>
```

**Benchmark a model** to measure response time and throughput:

```http
POST /api/admin/models/by-id/benchmark?model_id=<modelId>
Authorization: Bearer <access_token>
```

:::tip
Run a connection test after adding or updating a provider to catch misconfigured API keys or URLs early.
:::
