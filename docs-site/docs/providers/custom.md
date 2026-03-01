---
title: Custom OpenAI-Compatible Providers
sidebar_label: Custom Providers
sidebar_position: 6
---

# Custom OpenAI-Compatible Providers

Any API server that implements the OpenAI chat completions format (`/v1/chat/completions`) works with ClaraVerse. This covers a wide range of local inference engines, cloud platforms, and proxy services.

## Add a Custom Provider

1. Open the **Admin Panel > Providers**.
2. Click **Add Provider**.
3. Enter a name (e.g., "LM Studio", "Together AI").
4. Set the base URL to the API endpoint (e.g., `http://localhost:1234/v1`).
5. Enter an API key if the provider requires one, or leave it blank for local engines.
6. Click **Save**.

ClaraVerse will attempt to fetch the model list from the `/v1/models` endpoint. If that works, models appear automatically. If not, you can add model IDs manually.

## Compatible Services

### Local Inference

| Service | Typical Base URL | API Key | Notes |
|---------|-----------------|---------|-------|
| **LM Studio** | `http://localhost:1234/v1` | No | GUI-based, easy model management |
| **text-generation-webui** (oobabooga) | `http://localhost:5000/v1` | No | Supports many model formats |
| **vLLM** | `http://localhost:8000/v1` | No | High-throughput serving, production-grade |
| **llama.cpp server** | `http://localhost:8080/v1` | No | Minimal footprint |
| **LocalAI** | `http://localhost:8080/v1` | No | Drop-in OpenAI replacement |

### Cloud Platforms

| Service | Base URL | API Key | Notes |
|---------|----------|---------|-------|
| **Together AI** | `https://api.together.xyz/v1` | Yes | Wide model selection, fast inference |
| **Groq** | `https://api.groq.com/openai/v1` | Yes | Extremely fast inference (LPU hardware) |
| **Fireworks AI** | `https://api.fireworks.ai/inference/v1` | Yes | Fast, cost-effective |
| **Perplexity** | `https://api.perplexity.ai` | Yes | Search-augmented models |
| **Mistral AI** | `https://api.mistral.ai/v1` | Yes | Mistral and Mixtral models |
| **DeepSeek** | `https://api.deepseek.com/v1` | Yes | DeepSeek-R1, DeepSeek-V3 |
| **OpenRouter** | `https://openrouter.ai/api/v1` | Yes | Routes to 100+ models from one key |

## Setup Example: LM Studio

1. Open LM Studio and download a model from its model browser.
2. Go to the **Local Server** tab and click **Start Server**. Default port is `1234`.
3. In ClaraVerse, add a custom provider with base URL `http://localhost:1234/v1` and no API key.
4. The loaded model will appear in ClaraVerse's model selectors.

## Setup Example: Together AI

1. Sign up at [together.ai](https://www.together.ai/) and copy your API key from the dashboard.
2. In ClaraVerse, add a custom provider:
   - **Base URL:** `https://api.together.xyz/v1`
   - **API Key:** your Together AI key
3. Models like Llama 3.1 405B, Mixtral, and Code Llama will be available.

## Tips

- **Test the endpoint first.** Run a quick curl to make sure the API is reachable before adding it to ClaraVerse:
  ```bash
  curl http://localhost:1234/v1/models
  ```
- **Docker networking.** If both ClaraVerse and the inference server run in Docker, use the container name or `host.docker.internal` instead of `localhost`.
- **Model IDs.** Some providers use different model ID formats. If models don't auto-populate, check the provider's docs for the exact model ID string and add it manually.
- **Streaming.** ClaraVerse uses streaming by default. Most OpenAI-compatible APIs support this, but if you see issues, check if your provider requires a specific streaming parameter.
- **Multiple providers.** You can add the same service multiple times with different configurations -- for example, one entry for a local LM Studio instance and another for Together AI. Each gets its own name and model list.
