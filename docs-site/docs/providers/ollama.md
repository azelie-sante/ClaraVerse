---
title: Ollama (Local)
sidebar_label: Ollama
sidebar_position: 5
---

# Ollama (Local)

Run AI models locally on your own hardware with Ollama. No API key, no cloud dependency, completely free and private.

## Install Ollama

1. Go to [ollama.com](https://ollama.com/) and download the installer for your OS (macOS, Linux, or Windows).
2. Run the installer and follow the prompts.
3. Verify the installation:

```bash
ollama --version
```

Ollama runs a local API server on port `11434` by default.

## Pull a Model

Download at least one model before connecting to ClaraVerse:

```bash
# General-purpose
ollama pull llama3.1

# Smaller / faster
ollama pull llama3.1:8b

# Code-focused
ollama pull codellama

# Lightweight chat
ollama pull phi3
```

Browse all available models at [ollama.com/library](https://ollama.com/library).

## Connect Ollama to ClaraVerse

1. Open the **Admin Panel > Providers**.
2. Click **Add Provider**.
3. Enter a name (e.g., "Ollama").
4. Set the base URL to `http://localhost:11434/v1` (or your Ollama host's IP if running on a different machine).
5. Leave the API key field blank -- Ollama does not require one.
6. Click **Save**.

All models you have pulled will appear in model selectors automatically.

## Recommended Models

| Model | Size | Best For |
|-------|------|----------|
| **llama3.1:70b** | 40 GB | High-quality general tasks (needs beefy GPU) |
| **llama3.1:8b** | 4.7 GB | Good all-rounder for most hardware |
| **codellama:13b** | 7.4 GB | Code generation and debugging |
| **phi3:mini** | 2.3 GB | Fast responses on limited hardware |
| **mistral:7b** | 4.1 GB | Balanced performance and speed |
| **gemma2:9b** | 5.4 GB | Strong reasoning at small size |

Choose based on your available VRAM. As a rule of thumb, you need roughly 1 GB of VRAM per 1-2 billion parameters (quantized).

## Tips

- **Fully offline.** Once a model is pulled, it works without any internet connection. Ideal for air-gapped environments.
- **GPU acceleration.** Ollama automatically uses your GPU if available (NVIDIA, AMD, or Apple Silicon). Check `ollama ps` to confirm a model is loaded on the GPU.
- **Multiple models.** You can pull as many models as your disk can hold. Only the actively loaded model(s) consume VRAM.
- **Remote Ollama.** If Ollama runs on a different machine (e.g., a GPU server), set `OLLAMA_HOST=0.0.0.0` on that machine and point ClaraVerse to `http://<server-ip>:11434`.
- **Keep models updated.** Run `ollama pull <model>` again periodically to get the latest version of a model.
- **Docker networking.** If ClaraVerse and Ollama both run in Docker, use the Docker host IP or network alias instead of `localhost`. On Docker Desktop, `host.docker.internal` typically works.
