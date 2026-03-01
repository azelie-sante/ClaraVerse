---
title: What is ClaraVerse?
sidebar_label: Introduction
sidebar_position: 1
---

# What is ClaraVerse?

ClaraVerse is a self-hosted, private AI workspace. You run it on your own infrastructure, connect your own AI providers (OpenAI, Anthropic, Google, Ollama, or any OpenAI-compatible endpoint), and get a full-featured AI platform without sending data to third parties.

Built with a React frontend, Go backend, and Docker Compose deployment, ClaraVerse is designed to be up and running in minutes.

## Key Features

- **Chat** -- Multi-model conversations with file uploads, code execution, and web search built in.
- **Skills** -- 150+ built-in tools that activate mid-conversation -- web search, image generation, data analysis, and more.
- **Workflows** -- Visual drag-and-drop automation builder for chaining AI tasks.
- **Nexus** -- Assign Clara long-running tasks (research, coding, analysis). Track progress on a Kanban board.
- **Routines** -- Scheduled AI tasks that run automatically (cron-style) and report back via Telegram.
- **Channels** -- Telegram integration -- talk to Clara from your phone, receive routine reports.
- **Integrations** -- 150+ built-in integrations: Slack, GitHub, Jira, Google Sheets, Notion, Discord, Telegram, HubSpot, Shopify, and more. No MCP required.
- **Artifacts** -- AI-generated interactive content (charts, apps, visualizations) rendered inline.
- **AI Docs** -- Built-in tools for processing PPT, PDF, and CSV documents.
- **Memory** -- Clara remembers context across conversations and auto-archives old memories.
- **Devices** -- Connect all your machines -- Clara reaches MCP servers on any of them remotely.
- **Clara Companion** -- Bridge local MCP servers to ClaraVerse over WebSocket.
- **Local AI** -- Ollama and LM Studio auto-detected. Run models locally with zero configuration.
- **Browser-Local Storage** -- Conversations stored in IndexedDB. Zero-knowledge architecture.
- **BYOK** -- Bring Your Own Keys or use free local models.
- **Admin Panel** -- User management, provider configuration, model controls, and usage analytics.
- **Self-Hosted** -- Your data stays on your hardware. AGPL-3.0 licensed.

## Architecture at a Glance

| Component | Technology |
|-----------|-----------|
| Frontend | React 19 + TypeScript + Vite |
| Backend | Go (Fiber) |
| Databases | MongoDB (conversations, data), MySQL (providers, models), Redis (jobs, pub/sub) |
| Search | SearXNG (self-hosted, private) |
| Deployment | Docker Compose |

## Next Steps

Head to [Installation](./getting-started/installation.md) to get ClaraVerse running on your machine.
