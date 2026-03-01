---
title: "Why We Built ClaraVerse: The Self-Hosted AI Workspace That Started as a Joke"
slug: why-we-built-claraverse
authors: [badboysm890]
tags: [open-source, ai-workspace, self-hosted, privacy]
description: "How a joke about building a local AI role-playing app turned into ClaraVerse, the open source self-hosted AI workspace with chat, workflows, and zero-knowledge privacy."
keywords:
  - self-hosted AI workspace
  - open source AI assistant
  - private AI chatbot
  - self-hosted LLM
  - AI workflow builder
  - open source AI workspace
  - privacy-first AI
  - local AI tools
image: /img/image-banner.png
---

# Why We Built ClaraVerse: The Self-Hosted AI Workspace That Started as a Joke

I need to be honest with you. ClaraVerse started as a joke.

Not a "we had a grand vision for the future of AI" kind of story. More like a "I wanted to build a local role-playing app where you talk to AI and it shows images alongside the chat" kind of story. That is genuinely how this whole thing began.

<!-- truncate -->

## The Dumbest Idea That Turned Into Something Real

Back in early 2025, I (badboysm890) was messing around with local AI models. I had this silly idea: what if I could build an app where you chat with an AI character and it generates images in real time to go along with the conversation? Like a visual novel, but powered by LLMs and diffusion models running on your own machine.

So I started looking for tools that could do this. Chat with an AI? Sure, plenty of those. Generate images locally? Yep, got options. Combine both in one interface? Crickets.

Every single tool I found was laser-focused on **one thing**. You had your chat apps. You had your image generation UIs. You had your workflow automation tools. But nothing brought them together into a single, cohesive workspace. You were expected to juggle five different browser tabs, three different Docker containers, and a prayer.

That realization hit me harder than the original joke. The gap was not just about my silly role-playing idea. It was about the entire open source AI assistant ecosystem. Nobody had built the workspace. Everyone was building tools. There is a massive difference.

## From Joke to "Wait, This Could Actually Be Useful"

I pinged [aruntemme](https://github.com/aruntemme), and we started talking about what a real **self-hosted AI workspace** would look like. Not just another chatbot. Not just another Stable Diffusion frontend. A full workspace where AI is woven into everything you do.

We asked ourselves: what if you had one place where you could chat with any model, build visual automations, manage projects with AI teammates, and connect to all the services you already use? What if all of that ran on your own infrastructure, and your data never left your machine?

That question is what killed the joke and gave birth to ClaraVerse.

## What a Private AI Chatbot Should Actually Look Like

Most "private" AI tools still phone home. They store your conversations on someone else's server. They require accounts, API proxies, telemetry. ClaraVerse takes a fundamentally different approach.

We built a **zero-knowledge architecture**. Your conversations are stored in your browser's IndexedDB. The server never sees your chat history. There is no analytics pipeline slurping up your prompts. When you close the tab, your data stays exactly where you left it: on your machine, in your browser, under your control.

This is what a private AI chatbot should actually be. Not "we pinky-promise not to read your data." Instead: "we literally cannot read your data because we never receive it."

## Building the Open Source AI Assistant We Wanted to Use

Once we committed to building this thing properly, the feature list grew fast. Not because we were chasing hype, but because a real workspace needs real capabilities.

**[Chat with Skills](/docs/features/chat)** was the foundation. Talk to any model -- OpenAI, Anthropic, Google, Ollama, or any OpenAI-compatible endpoint -- and give it skills like web search, code execution, or file analysis. Multi-provider support was non-negotiable because vendor lock-in is the opposite of freedom.

**[Visual Workflow Builder](/docs/features/workflows)** came next. Drag-and-drop nodes to build AI automations without writing code. Connect an input to a model to a tool to an output. It is the kind of thing that costs $50/month on SaaS platforms, and we shipped it for free.

**[Nexus](/docs/features/nexus)** is probably my favorite feature. Think of it as a Trello-like board, except the AI is your team. You create cards, assign them to AI agents, and they work through the tasks. It sounds wild until you try it, and then you wonder how you ever managed projects without it.

We kept going. **150+ integrations** so ClaraVerse plugs into the tools you already use. **Telegram channels** so your AI can reach you where you are. **Routines** for scheduled automations that run without you babysitting them. An **MCP bridge** for connecting to the growing ecosystem of Model Context Protocol servers.

All open source. All self-hosted. All yours.

## Why Self-Hosted AI Workspaces Matter

There is a reason the self-hosted AI workspace category is exploding. People are tired of:

- Paying per-seat SaaS fees for AI tools that wrap the same APIs they could call directly
- Trusting third parties with sensitive business data just to use a chatbot
- Being locked into one model provider when the landscape changes every week
- Running six disconnected tools when one workspace could replace them all

ClaraVerse is our answer to all of that. You bring your own API keys or run local models with Ollama. You deploy on your own infrastructure with Docker Compose. You own every byte of data that flows through the system.

We are not building a product. We are building infrastructure for people who take their AI usage seriously and want it under their own roof.

## Community-Driven, Not VC-Driven

ClaraVerse is community-driven open source, licensed under AGPL-3.0. There is no venture capital pressure to enshittify the product. There is no pivot to enterprise-only features behind a paywall. The roadmap is driven by the people who actually use the tool and contribute to it.

aruntemme and I started this, but the community is what keeps it alive. Every issue filed, every PR merged, every feature request in Discord makes ClaraVerse better for everyone.

## Try It Yourself

If any of this resonates with you -- if you are tired of fragmented AI tools, if you want your conversations to stay private, if you want a real workspace instead of just another chat window -- give ClaraVerse a shot.

**Get started in under five minutes:**

1. Clone the repo: [github.com/claraverse/ClaraVerse-Scarlet-OSS](https://github.com/claraverse/ClaraVerse-Scarlet-OSS)
2. Run `docker compose up`
3. Open your browser and start building

Check out the [installation guide](/docs/getting-started/installation) for detailed setup instructions, or dive straight into the [feature docs](/docs/features/chat) to see what ClaraVerse can do.

Your private AI workspace is waiting. And yes, you can still use it for role-playing if you want. I will not judge.
