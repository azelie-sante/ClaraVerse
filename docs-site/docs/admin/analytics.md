---
title: Analytics
sidebar_label: Analytics
sidebar_position: 5
---

# Analytics

The analytics dashboard gives admins visibility into platform usage. Access it from **Admin Panel > Analytics**.

## Usage Overview

High-level stats across the entire instance:

```http
GET /api/admin/analytics/overview
Authorization: Bearer <access_token>
```

| Metric | Description |
|--------|-------------|
| `total_users` | All registered users |
| `active_chats` | Currently active chats |
| `total_messages` | Lifetime messages across the platform |
| `api_calls_today` | API calls made in the current day |
| `agent_executions` | Lifetime agent executions |
| `agents_run_today` | Agents run in the current day |

## Provider Analytics

See how each AI provider is being used:

```http
GET /api/admin/analytics/providers
Authorization: Bearer <access_token>
```

Per-provider metrics include:

- **Total requests** -- Lifetime API calls to this provider
- **Requests today** -- Calls made in the current day
- **Tokens used** -- Total token consumption
- **Avg response time** -- Mean latency in milliseconds

:::tip
Use provider analytics to identify cost hotspots. If one provider dominates token usage, consider adding a cheaper alternative for routine tasks.
:::

## Chat Analytics

Aggregated data on conversations across the platform:

```http
GET /api/admin/analytics/chats
Authorization: Bearer <access_token>
```

Tracks total chats, messages, and conversation patterns over time.

## Model Usage Stats

Breakdown by individual model:

```http
GET /api/admin/analytics/models
Authorization: Bearer <access_token>
```

| Metric | Description |
|--------|-------------|
| `total_requests` | Lifetime requests to this model |
| `requests_today` | Requests in the current day |
| `avg_tokens_per_request` | Average token count per call |
| `success_rate` | Percentage of successful responses |

Models with low success rates may indicate configuration issues or provider-side problems.

## Agent Execution Metrics

Track Nexus agent activity:

```http
GET /api/admin/analytics/agents
Authorization: Bearer <access_token>
```

Covers total agents created, execution counts, and trends. Useful for understanding how heavily the agent system is being used and which agents drive the most traffic.

## Background Jobs

ClaraVerse runs scheduled maintenance jobs that appear in backend logs:

| Job | Schedule | Purpose |
|-----|----------|---------|
| Retention Cleanup | Daily at 2:00 AM UTC | Deletes expired data per tier retention policy |
| Grace Period Check | Hourly | Handles tier downgrades |
| Promo Expiration | Hourly | Expires promotional tier upgrades |

Monitor these in your backend container logs:

```bash
docker logs claraverse-backend -f
```
