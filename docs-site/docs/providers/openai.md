---
title: OpenAI Setup
sidebar_label: OpenAI
sidebar_position: 2
---

# OpenAI Setup

Connect ClaraVerse to OpenAI to use GPT-4o, o1, o3, and other OpenAI models.

## Get an API Key

1. Go to [platform.openai.com](https://platform.openai.com/) and sign in (or create an account).
2. Navigate to **API keys** in the left sidebar (or visit [platform.openai.com/api-keys](https://platform.openai.com/api-keys)).
3. Click **Create new secret key**, give it a name, and copy the key. You will not be able to view it again.
4. Make sure you have billing enabled under **Settings > Billing** -- API access requires a paid account.

## Add OpenAI to ClaraVerse

1. Open the **Admin Panel > Providers**.
2. Click **Add Provider**.
3. Enter a name (e.g., "OpenAI").
4. Set the base URL to `https://api.openai.com/v1`.
5. Paste your API key.
6. Click **Save**.

Models will be fetched automatically and appear in model selectors across the platform.

## Available Models

| Model | Best For |
|-------|----------|
| **GPT-4o** | General-purpose, multimodal (text + vision) |
| **GPT-4o mini** | Fast, cost-effective tasks |
| **o1** | Complex reasoning, math, code |
| **o3** | Advanced reasoning with extended thinking |
| **o3-mini** | Lightweight reasoning |
| **GPT-4.1** | Long-context tasks (1M token window) |
| **GPT-4.1 mini** | Cost-effective long-context |

Model availability depends on your OpenAI account tier.

## Tips

- **Set spending limits.** Use OpenAI's usage limits page to cap monthly spend and avoid surprises.
- **Pick the right model.** GPT-4o mini handles most everyday tasks well and costs significantly less than GPT-4o. Reserve o1/o3 for problems that need deep reasoning.
- **Organization keys.** If you have an OpenAI organization, make sure to use an API key associated with that org so usage is billed correctly.
- **Rate limits.** OpenAI enforces per-minute token and request limits based on your usage tier. If users hit rate errors, consider upgrading your tier or spreading load across multiple keys.
