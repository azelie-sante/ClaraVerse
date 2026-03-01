---
title: Scheduled Automations
sidebar_label: Routines
sidebar_position: 4
---

# Routines -- Scheduled Automations

Routines are task sequences that run on a schedule. Define what Clara should do, set a time, and she handles the rest -- reporting results back to you via Telegram.

## Creating a Routine

1. Navigate to **Nexus** and click **Routines** in the Nexus sidebar.
2. Click **New Routine** inside the Routines view.
3. Describe the task sequence. For example:
   - "Every Monday at 9 AM, check open GitHub PRs, summarize them, and send the summary to my Telegram."
   - "Daily at 6 PM, pull today's analytics from Mixpanel and post a report to Slack #metrics."
4. Set the **schedule** (daily, weekly, custom cron expression).
5. Choose a **reporting channel** (Telegram).
6. Activate the routine.

## How Routines Differ from Workflows

| | Routines | Workflows |
|---|---|---|
| **Interface** | Text-based task description | Visual drag-and-drop canvas |
| **Best for** | Recurring AI tasks with natural language instructions | Complex multi-step automations with branching logic |
| **Execution** | Clara interprets and executes each run | Deterministic node-by-node execution |

Use **Routines** when you want Clara to figure out the steps. Use **Workflows** when you need precise control over every node.

## Key Capabilities

- **Flexible scheduling.** Pick from presets (hourly, daily, weekly) or write a cron expression for exact control.
- **Telegram reporting.** Results are pushed to your Telegram chat automatically. See [Channels](./channels.md) for setup.
- **Integration access.** Routines can use any of the 150+ built-in integrations -- pull data from APIs, write to databases, send messages.
- **Chained tasks.** A routine can contain multiple steps that execute in order, passing results from one to the next.

## Tips

- **Start simple.** Create a one-step routine ("Summarize today's Hacker News front page and send it to Telegram") to verify your schedule and reporting channel work before building multi-step sequences.
- **Combine with integrations.** Routines become powerful when they pull from services like Google Sheets, Jira, or GitHub and push results to Telegram or Slack.
- **Review history.** Each routine run is logged. Check past runs to see what Clara did and whether results met your expectations.
