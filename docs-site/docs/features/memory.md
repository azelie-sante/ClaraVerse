---
title: Memory System
sidebar_label: Memory
sidebar_position: 8
---

# Memory System

Clara remembers important context across conversations. When you share a preference, a project detail, or a fact you want Clara to recall later, it gets stored as a memory. Old or irrelevant memories are automatically archived so the active set stays useful.

## How It Works

- **Automatic capture** -- Clara identifies important facts during conversation and saves them as memories without you needing to do anything explicit.
- **Cross-conversation recall** -- Memories persist between chat sessions. Start a new conversation and Clara still knows your preferences and context.
- **Auto-archiving** -- Stale or outdated memories are automatically archived to keep the active memory set relevant.
- **Privacy-first** -- All memories are stored in your own database. Nothing leaves your ClaraVerse instance.

## Managing Memories

Memory is managed in **Settings > AI Configuration**.

From there you can:

- **Toggle memory on or off** to control whether Clara stores new memories.
- **Clear all memories** to reset everything Clara has remembered.

## What Gets Remembered

Clara stores things like:

- Your name, role, and preferences.
- Project names, tech stacks, and conventions you mention.
- Repeated instructions ("always use TypeScript", "I prefer dark mode").
- Facts you explicitly ask Clara to remember.

## Tips

- You can explicitly tell Clara to remember something: "Remember that our API uses v2 endpoints."
- If Clara recalls something outdated, tell it to forget: "Forget that I use Python -- I switched to Go."
- Memories are scoped to your user account. Other users on the same ClaraVerse instance have their own separate memory.
