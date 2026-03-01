---
title: Provider Management
sidebar_label: Provider Management
sidebar_position: 3
---

# Provider Management

Providers are the AI services that power ClaraVerse. Manage them from **Admin Panel > Providers**.

## Supported Provider Types

| Provider | Type value | API Key Required | Example Base URL |
|----------|-----------|-----------------|------------------|
| OpenAI | `openai` | Yes | `https://api.openai.com/v1` |
| Anthropic | `anthropic` | Yes | `https://api.anthropic.com` |
| Google AI | `google` | Yes | Default (Google endpoint) |
| Ollama | `openai` | No | `http://localhost:11434/v1` |
| Custom / OpenAI-compatible | `openai` | Varies | Your endpoint URL |

:::tip
Ollama and other local providers use the `openai` type because they expose an OpenAI-compatible API. No API key is needed for local providers.
:::

## Adding a Provider

1. Go to **Admin Panel > Providers**.
2. Click **Add Provider**.
3. Fill in the details:

```http
POST /api/admin/providers
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "OpenAI",
  "type": "openai",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-...",
  "enabled": true
}
```

For a local Ollama instance:

```json
{
  "name": "Local Ollama",
  "type": "openai",
  "base_url": "http://localhost:11434/v1",
  "api_key": "",
  "enabled": true
}
```

## Updating a Provider

Change the name, base URL, or API key:

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

## Enabling / Disabling Providers

Toggle a provider on or off without deleting it:

```http
PUT /api/admin/providers/:id/toggle
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "enabled": false
}
```

Disabled providers are hidden from users. Their models will not appear in model selectors.

## Fetching Models

After adding a provider, pull its available models:

```http
POST /api/admin/providers/:providerId/fetch
Authorization: Bearer <access_token>
```

This queries the provider's API and imports discovered models into ClaraVerse.

## Deleting a Provider

```http
DELETE /api/admin/providers/:id
Authorization: Bearer <access_token>
```

:::warning
Deleting a provider removes all its associated models. Disable the provider first and migrate any agents or workflows that depend on its models.
:::

## Admin Keys vs. User Keys

- **Admin key** -- Configured here in the Admin Panel. Shared by all users who have not added their own key for this provider.
- **User key** -- Set individually by users in **Settings > Providers**. Overrides the admin key for that user's requests.

If neither an admin key nor a user key is configured for a provider, requests to that provider will fail.
