---
title: Quickstart
sidebar_label: Quickstart
sidebar_position: 3
---

# Quickstart

You have ClaraVerse running. Here is what to do in your first 5 minutes.

## 1. Register Your Account

Open [http://localhost](http://localhost) and create an account. The **first user to register automatically becomes the admin** -- no invite code or special setup needed.

## 2. Add an AI Provider

ClaraVerse does not include AI models. You bring your own API keys.

1. Click the **gear icon** (Settings) in the sidebar.
2. Go to **Providers**.
3. Click **Add Provider** and choose one:
   - **OpenAI** -- Paste your API key from [platform.openai.com](https://platform.openai.com/api-keys).
   - **Anthropic** -- Paste your API key from [console.anthropic.com](https://console.anthropic.com/).
   - **Google AI** -- Paste your API key from [aistudio.google.com](https://aistudio.google.com/apikey).
   - **Ollama** -- Point to your Ollama instance (e.g., `http://host.docker.internal:11434` if running Ollama on the host machine). No API key needed.
   - **Custom** -- Any OpenAI-compatible API endpoint.
4. ClaraVerse will automatically detect available models from the provider.

## 3. Start Chatting

1. Click **Chat** in the sidebar.
2. Select a model from the dropdown at the top.
3. Type a message and press Enter.

You can upload files (drag and drop or click the attachment icon), and the AI can read PDFs, images, code files, and more.

## 4. Try a Skill

Skills are built-in tools the AI can use. Try web search:

1. In a chat, ask something that requires current information, for example: *"What are the top headlines today?"*
2. The AI will automatically use the **SearXNG web search** skill (already running as part of the Docker stack) to find and summarize results.

Other skills to try:
- Ask it to write and run code (requires E2B API key in your `.env`).
- Ask it to analyze a file you uploaded.

## 5. Explore Nexus

Nexus lets you create persistent AI agents with custom instructions and tool access.

1. Click **Nexus** in the sidebar.
2. Create a new agent -- give it a name, system instructions, and select which skills it can use.
3. Start a conversation with your agent. It remembers its role across sessions.

## What Else Can You Do?

- **Workflows** -- Build multi-step AI automations with a visual drag-and-drop editor.
- **Routines** -- Schedule AI tasks to run on a cron schedule (e.g., daily summaries, weekly reports).
- **Channels** -- Talk to Clara from Telegram when you're away from the app.
- **Settings > Admin** -- Manage users, configure which models are available, and view usage analytics.

## Next Steps

- [Configuration](./configuration.md) -- Customize environment variables and connect external services.
- [Updating](./updating.md) -- Keep ClaraVerse up to date.
