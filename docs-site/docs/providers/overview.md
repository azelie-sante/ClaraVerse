---
title: Multi-Provider Support
sidebar_label: Overview
sidebar_position: 1
---

# Multi-Provider Support

ClaraVerse connects to multiple AI providers at the same time. You bring your own API keys (BYOK), pick the models you want, and switch between them freely inside Chat, Workflows, Nexus, and Routines.

## How It Works

1. **Admin configures providers** -- An admin adds one or more AI providers in the Admin Panel and optionally sets a default API key for each.
2. **Users add their own keys** -- Individual users can add personal API keys in **Settings > Providers**. Personal keys override admin-provided ones.
3. **Models appear automatically** -- Once a provider is connected, its models show up in every model selector across the platform.

## Supported Providers

| Provider | Type | API Key Required | Notes |
|----------|------|-----------------|-------|
| [OpenAI](./openai.md) | Cloud | Yes | GPT-4o, o1, o3, and more |
| [Anthropic](./anthropic.md) | Cloud | Yes | Claude Opus 4, Sonnet 4, Haiku |
| [Google AI](./google.md) | Cloud | Yes | Gemini 2.5 Pro, Flash |
| [Ollama](./ollama.md) | Local | No | Free, fully offline |
| [Custom / OpenAI-compatible](./custom.md) | Varies | Varies | LM Studio, vLLM, Groq, Together AI, etc. |

## Admin vs. User Keys

- **Admin key** -- Set in the Admin Panel. Shared across all users who don't have their own key for that provider.
- **User key** -- Set in user settings. Takes priority over the admin key. Only that user's requests use it.

If no key is configured at either level, the provider is visible but requests will fail until a key is added.

## Adding a Provider (Quick Steps)

1. Go to **Admin Panel > Providers**.
2. Click **Add Provider**.
3. Enter a name (e.g., "OpenAI"), the base URL, and your API key.
4. Save. Models from that provider are now available to all users.

See the individual provider pages for detailed setup instructions.
