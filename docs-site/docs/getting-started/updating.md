---
title: Updating
sidebar_label: Updating
sidebar_position: 4
---

# Updating ClaraVerse

ClaraVerse is updated by pulling the latest code and rebuilding the Docker containers.

## Standard Update

```bash
# 1. Pull the latest changes
git pull

# 2. Stop the running stack
docker compose down

# 3. Rebuild and start
docker compose up --build -d
```

The rebuild picks up code changes in both the frontend and backend. Database volumes are preserved automatically -- your conversations, users, and settings carry over.

## Backup Before Updating

It is good practice to back up your data before major updates.

### Back Up the Environment File

```bash
cp .env .env.backup
```

### Back Up Docker Volumes

```bash
# Find your volume names
docker volume ls | grep claraverse

# Back up MongoDB data
docker run --rm -v claraverse_mongodb-data:/data -v $(pwd)/backups:/backup alpine \
  tar czf /backup/mongodb-backup-$(date +%Y%m%d).tar.gz -C /data .

# Back up MySQL data
docker run --rm -v claraverse_mysql-data-new:/data -v $(pwd)/backups:/backup alpine \
  tar czf /backup/mysql-backup-$(date +%Y%m%d).tar.gz -C /data .
```

### Back Up the Encryption Key

Your `ENCRYPTION_MASTER_KEY` in `.env` is critical. If you lose it, all encrypted data (conversations, saved credentials) becomes unrecoverable. Store a copy somewhere safe outside the server.

## Handling Breaking Changes

Occasionally, an update may include database migrations or new required environment variables. Check the release notes or changelog before updating. If new `.env` variables are added:

```bash
# Compare your .env with the latest template
diff .env .env.example
```

Add any missing variables to your `.env` file before restarting.

## Full Reset (Nuclear Option)

If something goes wrong and you want to start completely fresh:

```bash
# Stop everything and delete all volumes (THIS DELETES ALL DATA)
docker compose down -v

# Rebuild from scratch
docker compose up --build -d
```

:::danger
`docker compose down -v` removes all database volumes. You will lose all users, conversations, settings, and uploaded files. Only do this if you have backups or are okay losing everything.
:::
