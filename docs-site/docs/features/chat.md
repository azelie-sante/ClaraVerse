---
title: AI Chat with Skills
sidebar_label: Chat
sidebar_position: 1
---

# AI Chat with Skills

![Chat](/img/features/chat-demo.gif)

Clara is your AI assistant. Ask questions, get answers, and let Clara use **skills** -- context-aware tools that activate mid-conversation when needed. No manual toggling required.

## How It Works

1. Open **Chat** from the sidebar.
2. Pick a model (or let the default apply).
3. Type your message. Clara responds and activates skills automatically when they are relevant.

For example, ask "What happened in the news today?" and Clara triggers **web search** on its own. Ask "Generate a logo for my project" and **image generation** kicks in.

## Skills

![Skills](/img/features/skills.png)

Skills are tools Clara can call while chatting with you. They include:

- **Web Search** -- live results via the built-in SearXNG instance.
- **Image Generation** -- create images from text descriptions.
- **Data Analysis** -- process CSVs, run calculations, produce charts.
- **Code Execution** -- write and run code snippets in-conversation.
- **Document Tools** -- generate PDFs, PowerPoints, and more.

Skills are selected by the AI based on context. You do not need to enable them manually.

## Multi-Provider Support

Use any supported model provider:

| Provider | Examples |
|---|---|
| **OpenAI** | GPT-4o, GPT-4.1 |
| **Anthropic** | Claude Opus, Sonnet |
| **Google** | Gemini Pro, Flash |
| **Ollama** | Llama, Mistral, Qwen (auto-detected) |
| **Any OpenAI-compatible endpoint** | LM Studio, vLLM, etc. |

Switch providers per-conversation from the model selector.

## Human-in-the-Loop

Clara asks **visual clarifying questions** when she needs your input. Instead of plain text prompts, you get typed forms -- dropdowns, checkboxes, image selectors -- so you can answer precisely without guessing what format the AI expects.

## Conversation Storage

All conversations are stored **locally in your browser** using IndexedDB. Nothing leaves your machine unless you send a message to a cloud provider. This is a zero-knowledge architecture: the server never stores your chat history.

## Tips

- **Use integrations in chat.** Clara can pull data from Slack, GitHub, Jira, and 150+ other services directly in conversation. Set up credentials once in Settings > Integrations.
- **Artifacts.** Code, charts, and generated content appear as interactive artifacts you can revisit from the Artifacts panel.
- **Memory.** Clara remembers context across conversations and auto-archives old memories so she stays useful over time.
