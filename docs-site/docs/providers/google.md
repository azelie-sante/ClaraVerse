---
title: Google AI Setup
sidebar_label: Google AI
sidebar_position: 4
---

# Google AI Setup

Connect ClaraVerse to Google AI to use Gemini models for conversation, multimodal tasks, and long-context analysis.

## Get an API Key

1. Go to [Google AI Studio](https://aistudio.google.com/).
2. Sign in with your Google account.
3. Click **Get API key** in the top navigation (or visit [aistudio.google.com/apikey](https://aistudio.google.com/apikey)).
4. Click **Create API key**, select a Google Cloud project (or create one), and copy the generated key.

Google AI Studio provides free-tier access with rate limits. For higher throughput, enable billing on the associated Google Cloud project.

## Add Google AI to ClaraVerse

1. Open the **Admin Panel > Providers**.
2. Click **Add Provider**.
3. Enter a name (e.g., "Google AI").
4. Set the base URL to `https://generativelanguage.googleapis.com/v1beta`.
5. Paste your API key.
6. Click **Save**.

Gemini models will appear in model selectors across the platform.

## Available Models

| Model | Best For |
|-------|----------|
| **Gemini 2.5 Pro** | Complex reasoning, coding, long-context analysis |
| **Gemini 2.5 Flash** | Fast, cost-effective everyday tasks |
| **Gemini 2.0 Flash** | Lightweight tasks, high-speed responses |

Gemini models support multimodal inputs (text, images, video, audio) and context windows up to 1M tokens.

## Tips

- **Free tier.** Google AI Studio offers a generous free tier with per-minute rate limits. This is a great way to get started without a credit card.
- **Flash for most tasks.** Gemini 2.5 Flash handles the majority of use cases well and responds quickly. Use 2.5 Pro when you need deeper reasoning or very long context.
- **Multimodal inputs.** Gemini natively understands images, PDFs, and audio. Upload files in Chat and pick a Gemini model to take advantage of this.
- **Vertex AI.** If you need enterprise-grade SLAs or are already on Google Cloud, you can point ClaraVerse at a Vertex AI endpoint instead. Set the base URL to your Vertex AI endpoint and use a service account key.
- **Rate limits.** Free-tier keys are rate-limited (typically 15 requests per minute for Pro, 30 for Flash). If you hit limits, enable billing on your Google Cloud project to unlock higher quotas.
