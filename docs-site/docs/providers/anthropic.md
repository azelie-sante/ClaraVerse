---
title: Anthropic Setup
sidebar_label: Anthropic
sidebar_position: 3
---

# Anthropic Setup

Connect ClaraVerse to Anthropic to use Claude models for conversation, analysis, and code generation.

## Get an API Key

1. Go to [console.anthropic.com](https://console.anthropic.com/) and sign in (or create an account).
2. Navigate to **API Keys** in the dashboard.
3. Click **Create Key**, name it, and copy the key immediately -- it is shown only once.
4. Add a payment method under **Billing** if you haven't already. API access requires an active billing plan.

## Add Anthropic to ClaraVerse

1. Open the **Admin Panel > Providers**.
2. Click **Add Provider**.
3. Enter a name (e.g., "Anthropic").
4. Set the base URL to `https://api.anthropic.com`.
5. Paste your API key.
6. Click **Save**.

Anthropic models will appear in all model selectors once the provider is saved.

## Available Models

| Model | Best For |
|-------|----------|
| **Claude Opus 4** | Complex analysis, nuanced writing, deep reasoning |
| **Claude Sonnet 4** | Balanced performance and speed for most tasks |
| **Claude Haiku** | Fast, lightweight responses, high throughput |

All Claude models support large context windows (up to 200K tokens) and work with text, code, and vision inputs.

## Tips

- **Start with Sonnet.** Claude Sonnet 4 is the best all-around choice for most workloads -- fast, capable, and cost-effective.
- **Use Opus for hard problems.** Reserve Claude Opus 4 for tasks that require deep reasoning, complex code generation, or long-form analysis.
- **Haiku for speed.** When latency matters more than depth (quick classifications, short answers, routing), Haiku is the fastest option.
- **Workspace billing.** If you are part of an Anthropic workspace, make sure the API key belongs to the correct workspace so billing is attributed properly.
- **Rate limits.** Anthropic applies rate limits per model. If you see 429 errors, check your tier limits at [console.anthropic.com](https://console.anthropic.com/) and consider requesting a limit increase.
