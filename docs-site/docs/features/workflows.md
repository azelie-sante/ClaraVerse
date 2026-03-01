---
title: Visual Workflow Builder
sidebar_label: Workflows
sidebar_position: 2
---

# Visual Workflow Builder

![Workflows](/img/features/workflows-screenshot.png)

Build automations by connecting nodes on a canvas. Drag, drop, wire up -- no code required. Workflows support parallel execution, scheduling, and 150+ integrations.

## Getting Started

1. Open **Agents** from the sidebar (or the Workflows card on the Dashboard).
2. Click the **+** button to create a new agent workflow.
3. Drag nodes from the panel onto the canvas.
4. Connect outputs to inputs by drawing edges between nodes.
5. Click **Run** to execute.

## AI-Assisted Generation

Describe what you want in plain text and let the LLM build the workflow for you. For example:

> "Every morning, pull open GitHub issues labeled 'bug', summarize them, and post the summary to Slack #engineering."

Clara generates the full node graph. Review it, tweak if needed, and activate.

## Key Capabilities

- **Parallel execution.** Branches run concurrently. A workflow that fans out to five APIs finishes as fast as the slowest branch, not the sum of all five.
- **Scheduling.** Set workflows to run on a cron schedule -- hourly, daily, weekly, or custom.
- **API triggers.** Expose a workflow as an HTTP endpoint to call from external apps or webhooks.
- **150+ integrations.** Use the same integrations available in Chat and Nexus -- Slack, GitHub, Google Sheets, databases, and more.
- **Version history.** Every save creates a version. Roll back to any previous state.
- **Conditional logic.** Branch based on data values, API responses, or AI decisions.

## Node Types

| Type | What it does |
|---|---|
| **Trigger** | Starts the workflow (manual, schedule, webhook, or event) |
| **AI / LLM** | Sends a prompt to a model and returns the response |
| **Integration** | Calls an external service (Slack, GitHub, etc.) |
| **Transform** | Reshapes data -- map, filter, merge, split |
| **Condition** | Routes execution based on a true/false check |
| **Code** | Runs custom code for anything the built-in nodes do not cover |

## Tips

- **Start with AI generation**, then refine. It is faster than building from scratch.
