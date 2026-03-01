---
title: User Management
sidebar_label: User Management
sidebar_position: 2
---

# User Management

Manage every user on your ClaraVerse instance from **Admin Panel > Users**.

## Viewing Users

The user list supports searching and filtering:

| Filter | Options |
|--------|---------|
| **Search** | By user ID or domain |
| **Tier** | `free`, `pro`, `max`, `enterprise` |

Results are paginated (default 50 per page).

```http
GET /api/admin/users?page=1&page_size=50&search=john&tier=pro
Authorization: Bearer <access_token>
```

Each user entry shows User ID (anonymized), Domain, Tier, Chats, Messages, Agents, Last Active, and Created. Data is GDPR-compliant and anonymized -- full email addresses are not displayed.

## User Details

- **Limits** -- The tier-based defaults (messages per day, executions per day, retention days)
- **Overrides** -- Any admin-applied adjustments to those limits
- **Usage** -- Current consumption (messages today, executions today)

```http
GET /api/admin/users/:userID
Authorization: Bearer <access_token>
```

## Setting Usage Limits

Override tier defaults for individual users. This is useful for power users who need higher limits or accounts that should be restricted.

```http
POST /api/admin/users/:userID/overrides
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "limits": {
    "messages_per_day": 2000,
    "executions_per_day": 200
  },
  "reason": "Power user needs higher limits"
}
```

You can also override a user's tier directly:

```http
POST /api/admin/users/:userID/overrides
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "tier": "pro",
  "reason": "Upgraded for trial period"
}
```

To remove all overrides and revert to tier defaults:

```http
DELETE /api/admin/users/:userID/overrides
Authorization: Bearer <access_token>
```

:::tip
Document the reason for any override. Increased limits can affect resource consumption and API costs.
:::

## User Roles

ClaraVerse has two roles:

| Role | Access |
|------|--------|
| `user` | Standard platform access |
| `admin` | Full platform access plus Admin Panel |

### Promoting a User to Admin

Set the role directly in MongoDB:

```javascript
db.users.updateOne(
  { email: "user@example.com" },
  { $set: { role: "admin" } }
)
```

### SUPERADMIN_USER_IDS

For additional admin access without modifying the database, add user IDs to the `SUPERADMIN_USER_IDS` environment variable in your `.env` file:

```bash
SUPERADMIN_USER_IDS=user-id-1,user-id-2,user-id-3
```

:::warning
Keep the list of superadmin IDs short. Every admin account is a privileged access point.
:::
