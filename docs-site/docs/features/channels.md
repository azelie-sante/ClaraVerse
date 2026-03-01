---
title: Telegram Integration
sidebar_label: Channels
sidebar_position: 5
---

# Channels -- Telegram Integration

![Channels](/img/features/channels.png)

Channels let you talk to Clara from Telegram when you are away from the ClaraVerse app. You can also receive routine and workflow reports directly on your phone.

## Setup

### 1. Create a Telegram Bot

1. Open Telegram and search for **@BotFather**.
2. Send `/newbot` and follow the prompts to name your bot.
3. BotFather gives you a **bot token** (a string like `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`). Copy it.

### 2. Connect to ClaraVerse

1. In ClaraVerse, go to **Settings > Channels**.
2. Paste your Telegram bot token.
3. Click **Connect**.
4. Open your new bot in Telegram and send `/start`.

ClaraVerse confirms the connection. You can now message Clara directly in Telegram.

## What You Can Do

- **Chat with Clara.** Send messages in Telegram and get AI responses, just like in the web app.
- **Receive routine reports.** When a Routine completes, its results are sent to your Telegram chat automatically.
- **Get notifications.** Nexus task completions and other events can be routed to Telegram.

## Key Details

- **One bot per user.** Each ClaraVerse user connects their own Telegram bot for privacy.
- **Same skills and integrations.** Clara has access to the same tools in Telegram as in the web app -- web search, integrations, and more.
- **Media support.** Clara can send images, files, and formatted text in Telegram responses.

## Tips

- **Pin your bot** in Telegram for quick access.
- **Use Routines + Channels together** for a daily briefing delivered to your phone each morning.
- **Security.** Your bot token is stored on your ClaraVerse server. If you self-host, only you have access to it.
