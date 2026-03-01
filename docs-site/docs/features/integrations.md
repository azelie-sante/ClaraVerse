---
title: 150+ Built-in Integrations
sidebar_label: Integrations
sidebar_position: 6
---

# 150+ Built-in Integrations

![Integrations](/img/features/integrations.png)

ClaraVerse ships with 150+ integrations out of the box. Connect to the services you already use -- no MCP servers, no plugins, no extra setup beyond entering your credentials.

## Available Integrations

A partial list of supported services:

| Category | Services |
|---|---|
| **Communication** | Slack, Discord, Telegram, Microsoft Teams |
| **Project Management** | Jira, Linear, Asana, Trello, Notion |
| **Development** | GitHub, GitLab, Bitbucket |
| **Productivity** | Google Sheets, Google Drive, Google Calendar, Airtable |
| **CRM & Marketing** | HubSpot, Salesforce, Mailchimp |
| **E-Commerce** | Shopify, Stripe |
| **Analytics** | Mixpanel, PostHog, Google Analytics |
| **Cloud & Infra** | AWS S3, Google Cloud Storage |
| **Databases** | MongoDB, Redis, PostgreSQL, MySQL |
| **Other** | Twilio, SendGrid, Zapier Webhooks, custom HTTP |

This is not exhaustive. Check **Settings > Integrations** in the app for the full list.

## Setting Up Credentials

1. Go to **Settings > Integrations**.
2. Find the service you want to connect.
3. Enter the required credentials (API key, OAuth token, or connection string depending on the service).
4. Click **Save**.

That is it. The integration is now available across all ClaraVerse features.

## Shared Across Everything

Once configured, integrations are available in:

- **Chat** -- Clara can pull data from connected services mid-conversation.
- **Workflows** -- use integration nodes in your visual automations.
- **Nexus** -- Clara accesses integrations while working on long-running tasks.
- **Routines** -- scheduled tasks can read from and write to any connected service.

Connect once, use everywhere.

## No MCP Required

These integrations are **built in**. You do not need to run MCP servers or install external tools to use them. They work immediately after adding credentials.

If you do want to use MCP for additional tools or local filesystem access, ClaraVerse supports that too via [Clara Companion](./companion.md) -- but it is entirely optional.

## Tips

- **Principle of least privilege.** When creating API keys for integrations, scope them to the minimum permissions Clara needs.
- **Test after connecting.** Open Chat and ask Clara something that uses the integration (e.g., "List my open GitHub issues") to verify the connection works.
- **Multiple accounts.** You can connect multiple accounts for the same service if needed (e.g., two different GitHub organizations).
