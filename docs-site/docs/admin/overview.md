---
title: Admin Panel Overview
sidebar_label: Overview
sidebar_position: 1
---

# Admin Panel Overview

The Admin Panel gives instance administrators full control over users, AI providers, models, and platform analytics. Access it at `/admin` in your ClaraVerse deployment.

## Becoming an Admin

The first user to register automatically becomes the admin. There is no default admin account.

### Adding More Admins

**Option 1 -- Set the role in MongoDB (recommended):**

```javascript
db.users.updateOne(
  { email: "admin@example.com" },
  { $set: { role: "admin" } }
)
```

**Option 2 -- Environment variable:**

```bash
# .env
SUPERADMIN_USER_IDS=user-id-1,user-id-2
```

## Accessing the Admin Panel

| Environment | URL |
|-------------|-----|
| Development | `http://localhost:5173/admin` |
| Production  | `https://yourdomain.com/admin` |

Only users with the `admin` role or whose IDs appear in `SUPERADMIN_USER_IDS` can access the panel.

## What Admins Can Do

| Area | Capabilities |
|------|-------------|
| **[User Management](./user-management.md)** | View all users, set usage limits and overrides, assign roles |
| **[Provider Management](./provider-management.md)** | Add and configure AI providers, manage API keys, enable/disable providers |
| **[Model Management](./model-management.md)** | Create model entries, set capabilities, manage aliases, test and benchmark |
| **[Analytics](./analytics.md)** | View usage stats, token consumption, provider performance, agent metrics |

## Dashboard

The admin dashboard displays system-wide stats at a glance:

- Total Users
- Active Chats
- Total Messages
- API Calls Today
- Active Providers
- Total Models

You can verify your admin status programmatically:

```http
GET /api/admin/me
Authorization: Bearer <access_token>
```

Returns `"is_admin": true` if authenticated as an admin.

## System Health

Check that all services are running:

```http
GET /health
```

```json
{
  "status": "...",
  "connections": { ... },
  "timestamp": "..."
}
```
