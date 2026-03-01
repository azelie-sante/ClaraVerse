---
title: Local AI
sidebar_label: Local AI
sidebar_position: 12
---

# Local AI

Run AI models locally on your own hardware with Ollama or LM Studio. No API keys, no cloud dependency, completely free and private. You can also mix local and cloud providers for a hybrid setup.

## Supported Local Providers

| Provider | Type | How to Connect |
|----------|------|---------------|
| **Ollama** | Local inference | Add as provider with base URL `http://localhost:11434/v1` |
| **LM Studio** | Local inference (GUI) | Add as provider with base URL `http://localhost:1234/v1` |
| **Any OpenAI-compatible server** | Varies | Add as provider with the server's base URL |

## Setting Up Ollama

1. Install Ollama from [ollama.com](https://ollama.com/).
2. Pull a model: `ollama pull llama3.1`
3. In ClaraVerse, go to **Admin Panel > Providers**, click **Add Provider**.
4. Enter name "Ollama", base URL `http://localhost:11434/v1`, leave API key blank.
5. Save. Your pulled models appear in model selectors.

**If ClaraVerse runs in Docker**, Ollama on `localhost` won't be reachable from inside the container. Fix this by setting Ollama to listen on all interfaces:

```bash
# Linux (systemd):
sudo systemctl edit ollama
```

Add under `[Service]`:

```ini
Environment="OLLAMA_HOST=0.0.0.0"
```

Then restart:

```bash
sudo systemctl restart ollama
```

Use `http://host.docker.internal:11434/v1` (Docker Desktop) or your machine's IP as the base URL in ClaraVerse.

On macOS, set the environment variable before starting:

```bash
OLLAMA_HOST=0.0.0.0 ollama serve
```

## Setting Up LM Studio

1. Open LM Studio and download the models you want.
2. Start the local server from the LM Studio interface (runs on port 1234 by default).
3. In ClaraVerse, add a provider with name "LM Studio", base URL `http://localhost:1234/v1`, no API key.
4. The loaded model appears in ClaraVerse's model selectors.

## BYOK -- Bring Your Own Keys

ClaraVerse supports a hybrid approach:

- **Free local models** -- Use Ollama or LM Studio models at no cost.
- **Cloud providers** -- Add your own API keys for OpenAI, Anthropic, Google, or any OpenAI-compatible endpoint.
- **Mix and match** -- Use a local model for everyday tasks and a cloud model for specialized work.

You are never locked into a single provider. Switch models per conversation or per task.

## Running Fully Offline

To run ClaraVerse with no internet dependency:

1. Install ClaraVerse via Docker (you need internet for the initial pull).
2. Install Ollama and pull the models you want while online.
3. Set `OLLAMA_HOST=0.0.0.0` so ClaraVerse can reach Ollama from Docker.
4. Disconnect from the internet. ClaraVerse and Ollama work entirely locally.

All conversations are stored in your local database. Nothing is sent externally.

## Tips

- Local models run on your hardware. Ensure you have sufficient RAM and (optionally) a GPU. 8 GB RAM is the minimum; 16 GB or more is recommended for larger models.
- GPU acceleration is automatic with Ollama if you have a supported GPU (NVIDIA, AMD, or Apple Silicon). Run `ollama ps` to verify a model is loaded on the GPU.
- You can pull as many Ollama models as your disk can hold. Only the actively loaded model(s) consume VRAM.
- If you pull new models in Ollama, refresh the provider in ClaraVerse to pick them up.
