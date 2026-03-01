# ClaraVerse Quick Reference

Quick reference for common commands, environment variables, and troubleshooting.

## Quick Start

```bash
# Clone and setup
git clone https://github.com/claraverse-space/ClaraVerseAI.git
cd ClaraVerseAI
cp .env.example .env

# Generate secrets
echo "ENCRYPTION_MASTER_KEY=$(openssl rand -hex 32)" >> .env
echo "JWT_SECRET=$(openssl rand -hex 64)" >> .env

# Start everything
docker compose up -d

# Access
# Frontend: http://localhost:5173
# Backend:  http://localhost:3001
# Admin:    admin@localhost / admin
```

---

## Docker Commands

| Command | Description |
|---------|-------------|
| `docker compose up -d` | Start all services |
| `docker compose up -d --build` | Rebuild and start |
| `docker compose down` | Stop all services |
| `docker compose down -v` | Stop and remove volumes |
| `docker compose logs -f backend` | Follow backend logs |
| `docker compose logs -f frontend` | Follow frontend logs |
| `docker compose restart backend` | Restart backend only |
| `docker compose ps` | List running services |
| `docker compose exec backend sh` | Shell into backend |

---

## Development Commands

### Backend (Go)

```bash
cd backend

# Run server
go run cmd/server/main.go

# Run with hot reload (requires air)
air

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build binary
go build -o bin/server cmd/server/main.go

# Check for issues
go vet ./...
go mod tidy
```

### Frontend (React)

```bash
cd frontend

# Start dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type checking
npm run typecheck

# Linting
npm run lint
npm run lint:fix

# Format code
npm run format
```

### E2B Service (Python)

```bash
cd backend/e2b-service

# Create virtual environment
python -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run service
python main.py
```

---

## Environment Variables

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `ENCRYPTION_MASTER_KEY` | Data encryption key | `openssl rand -hex 32` |
| `JWT_SECRET` | JWT signing secret | `openssl rand -hex 64` |

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `development` | Environment mode |
| `PORT` | `3001` | Backend port |
| `ALLOWED_ORIGINS` | `http://localhost:5173` | CORS origins |
| `FRONTEND_URL` | `http://localhost:5173` | Frontend URL |
| `BACKEND_URL` | `http://localhost:3001` | Backend URL |

### Database

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_URI` | `mongodb://mongodb:27017/claraverse` | MongoDB connection |
| `MYSQL_PASSWORD` | `claraverse_pass_2024` | MySQL password |
| `REDIS_URL` | `redis://redis:6379` | Redis connection |

### Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_GLOBAL_API` | `200` | Global API limit/min |
| `RATE_LIMIT_PUBLIC_READ` | `120` | Public endpoints/min |
| `RATE_LIMIT_AUTHENTICATED` | `60` | Auth requests/min |
| `RATE_LIMIT_WEBSOCKET` | `20` | WS connections/min |
| `RATE_LIMIT_IMAGE_PROXY` | `60` | Image proxy/min |

### Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_ACCESS_TOKEN_EXPIRY` | `15m` | Access token lifetime |
| `JWT_REFRESH_TOKEN_EXPIRY` | `168h` | Refresh token lifetime (7 days) |
| `SUPERADMIN_USER_IDS` | (empty) | Comma-separated admin user IDs |

### External Services

| Variable | Description |
|----------|-------------|
| `SEARXNG_URLS` | SearXNG instance URLs |
| `E2B_API_KEY` | E2B cloud API key (optional) |
| `COMPOSIO_API_KEY` | Composio integration key |

---

## API Endpoints Quick Reference

### Authentication

```bash
# Register
curl -X POST http://localhost:3001/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","name":"User"}'

# Login
curl -X POST http://localhost:3001/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Get current user
curl http://localhost:3001/api/auth/me \
  -H "Authorization: Bearer <token>"
```

### Agents

```bash
# List agents
curl http://localhost:3001/api/agents \
  -H "Authorization: Bearer <token>"

# Create agent
curl -X POST http://localhost:3001/api/agents \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Agent","model":"gpt-4o","tools":["search_web"]}'
```

### Admin

```bash
# List users
curl http://localhost:3001/api/admin/users \
  -H "Authorization: Bearer <admin_token>"

# Get analytics
curl http://localhost:3001/api/admin/analytics/overview \
  -H "Authorization: Bearer <admin_token>"
```

### Health Check

```bash
curl http://localhost:3001/health
```

---

## Database Commands

### MongoDB

```bash
# Connect to MongoDB shell
docker exec -it claraverse-mongodb mongosh claraverse

# Common queries
db.users.find().count()
db.users.findOne({ email: "user@example.com" })
db.agents.find({ user_id: "user_abc123" })
db.executions.find().sort({ created_at: -1 }).limit(10)

# Make user admin
db.users.updateOne(
  { email: "user@example.com" },
  { $set: { role: "admin" } }
)
```

### MySQL

```bash
# Connect to MySQL shell
docker exec -it claraverse-mysql mysql -u claraverse_user -p claraverse

# Common queries
SELECT * FROM providers;
SELECT * FROM models WHERE visible = 1;
SELECT COUNT(*) FROM models GROUP BY provider_id;
```

### Redis

```bash
# Connect to Redis CLI
docker exec -it claraverse-redis redis-cli

# Common commands
KEYS *
KEYS "scheduler:*"
GET "key_name"
DEL "key_name"
FLUSHALL  # Clear all (careful!)
```

---

## WebSocket Testing

```bash
# Using websocat
websocat "ws://localhost:3001/ws/chat?token=<access_token>"

# Send message
{"type":"chat","payload":{"message":"Hello","model":"gpt-4o"}}
```

---

## Logs & Debugging

### View Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f backend
docker compose logs -f frontend

# Filter errors
docker compose logs backend 2>&1 | grep -E "ERROR|FATAL"

# Security events
docker compose logs backend 2>&1 | grep -E "🚫|⚠️"
```

### Enable Debug Mode

```bash
# In .env
LOG_LEVEL=debug
```

---

## Troubleshooting

### Port Already in Use

```bash
# Find process
lsof -i :3001
lsof -i :5173

# Kill process
kill -9 <PID>
```

### Database Connection Failed

```bash
# Check containers
docker compose ps

# Check logs
docker compose logs mongodb
docker compose logs mysql

# Restart databases
docker compose restart mongodb mysql redis
```

### Clear Rate Limits

```bash
# Clear specific user
docker exec -it claraverse-redis redis-cli DEL "ratelimit:user_abc123"

# Clear all rate limits
docker exec -it claraverse-redis redis-cli KEYS "ratelimit:*" | xargs redis-cli DEL
```

### Reset Everything

```bash
# Stop and remove all data
docker compose down -v

# Remove build cache
docker builder prune -f

# Fresh start
docker compose up -d --build
```

### Module/Dependency Issues

```bash
# Go
cd backend
go clean -modcache
go mod download
go mod tidy

# Node
cd frontend
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

---

## Service Ports

| Service | Port | URL |
|---------|------|-----|
| Frontend | 5173 (dev) / 80 (docker) | http://localhost:5173 |
| Backend | 3001 | http://localhost:3001 |
| MongoDB | 27017 | mongodb://localhost:27017 |
| MySQL | 3306 | mysql://localhost:3306 |
| Redis | 6379 | redis://localhost:6379 |
| SearXNG | 8080 (internal) | - |
| E2B | 8001 (internal) | - |

---

## File Locations

| File/Directory | Description |
|----------------|-------------|
| `.env` | Environment configuration |
| `docker-compose.yml` | Service orchestration |
| `backend/cmd/server/main.go` | Backend entry point |
| `backend/internal/` | Backend source code |
| `backend/migrations/` | Database migrations |
| `frontend/src/` | Frontend source code |
| `searxng/settings.yml` | SearXNG configuration |

---

## Useful Links

| Resource | URL |
|----------|-----|
| GitHub | https://github.com/claraverse-space/ClaraVerseAI |
| Discord | https://discord.com/invite/j633fsrAne |
| Website | https://claraverse.space |
| Cloud App | https://claraverse.app |

---

## Related Documentation

- [Architecture Guide](ARCHITECTURE.md) - System design
- [API Reference](API_REFERENCE.md) - Full API documentation
- [Developer Guide](DEVELOPER_GUIDE.md) - Development setup
- [Security Guide](FINAL_SECURITY_INSPECTION.md) - Security details
- [Admin Guide](ADMIN_GUIDE.md) - Administration
