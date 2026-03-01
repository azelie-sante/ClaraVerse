# ClaraVerse Developer Guide

Complete guide for setting up a local development environment for ClaraVerseAI.

## Prerequisites

### Required Software

| Software | Version | Purpose |
|----------|---------|---------|
| **Go** | 1.24+ | Backend API server |
| **Node.js** | 20+ | Frontend development |
| **Python** | 3.11+ | E2B code execution service |
| **Docker** | Latest | Database services |
| **tmux** | Latest | Multi-pane development (optional) |

### Installation

**macOS:**
```bash
# Using Homebrew
brew install go node python@3.11 docker tmux
```

**Ubuntu/Debian:**
```bash
# Go
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Node.js (via nvm)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 20

# Python
sudo apt install python3.11 python3.11-venv

# Docker
sudo apt install docker.io docker-compose-plugin
sudo usermod -aG docker $USER

# tmux
sudo apt install tmux
```

**Windows (WSL2 recommended):**
```powershell
# Install WSL2, then follow Ubuntu instructions
wsl --install -d Ubuntu
```

---

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/claraverse-space/ClaraVerseAI.git
cd ClaraVerseAI
```

### 2. Environment Setup

```bash
# Copy example environment file
cp .env.example .env

# Generate required secrets
echo "ENCRYPTION_MASTER_KEY=$(openssl rand -hex 32)" >> .env
echo "JWT_SECRET=$(openssl rand -hex 64)" >> .env
```

### 3. Install Dependencies

```bash
# Using Makefile (recommended)
make install

# Or manually:
cd backend && go mod download
cd ../frontend && npm install
cd ../backend/e2b-service && pip install -r requirements.txt
```

### 4. Start Development

**Option A: Using dev.sh (tmux)**

```bash
./dev.sh
```

This opens a 4-pane tmux session:
- Top-left: Backend (Go)
- Top-right: Frontend (React)
- Bottom-left: E2B Service (Python)
- Bottom-right: Info panel

**Option B: Manual startup**

```bash
# Terminal 1: Start databases
docker compose up mongodb mysql redis searxng -d

# Terminal 2: Start backend
cd backend
go run cmd/server/main.go

# Terminal 3: Start frontend
cd frontend
npm run dev

# Terminal 4: Start E2B service
cd backend/e2b-service
python main.py
```

### 5. Access Application

| Service | URL |
|---------|-----|
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:3001 |
| API Health | http://localhost:3001/health |

**Default Admin:**
```
Email: admin@localhost
Password: admin
```

---

## Project Structure

```
ClaraVerseAI/
├── backend/
│   ├── cmd/server/          # Main entry point
│   ├── internal/
│   │   ├── config/          # Configuration
│   │   ├── handlers/        # HTTP handlers
│   │   ├── middleware/      # Auth, rate limiting
│   │   ├── models/          # Data models
│   │   ├── services/        # Business logic
│   │   └── tools/           # 80+ tool implementations
│   ├── pkg/
│   │   ├── auth/            # JWT authentication
│   │   └── database/        # DB connections
│   ├── e2b-service/         # Python code execution
│   ├── migrations/          # MySQL migrations
│   └── docs/                # Documentation
├── frontend/
│   ├── src/
│   │   ├── components/      # React components
│   │   ├── stores/          # Zustand stores
│   │   ├── hooks/           # Custom hooks
│   │   └── utils/           # Utilities
│   └── public/              # Static assets
├── searxng/                 # Search engine config
├── docker-compose.yml       # Service orchestration
└── .env.example             # Environment template
```

---

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```bash
# Environment Mode
ENVIRONMENT=development

# Frontend URLs
VITE_API_BASE_URL=http://localhost:3001
VITE_WS_URL=ws://localhost:3001

# Backend
ALLOWED_ORIGINS=http://localhost,http://localhost:5173,http://localhost:5174
FRONTEND_URL=http://localhost:5173
BACKEND_URL=http://localhost:3001

# Security (REQUIRED - generate these!)
ENCRYPTION_MASTER_KEY=<openssl rand -hex 32>
JWT_SECRET=<openssl rand -hex 64>

# Database
MYSQL_ROOT_PASSWORD=claraverse_root_2024
MYSQL_PASSWORD=claraverse_pass_2024
MONGODB_URI=mongodb://localhost:27017/claraverse
REDIS_URL=redis://localhost:6379

# Rate Limiting
RATE_LIMIT_GLOBAL_API=200
RATE_LIMIT_WEBSOCKET=20

# Admin (optional)
SUPERADMIN_USER_IDS=
```

### Database Configuration

**Development (Docker):**
```yaml
# docker-compose.yml automatically configures:
MongoDB: mongodb://localhost:27017/claraverse
MySQL: mysql://claraverse_user:claraverse_pass_2024@localhost:3306/claraverse
Redis: redis://localhost:6379
```

**External databases:**
```bash
# MongoDB Atlas
MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/claraverse

# Managed MySQL
DATABASE_URL=mysql://user:pass@host:3306/claraverse

# Redis Cloud
REDIS_URL=redis://user:pass@host:6379
```

---

## Development Workflows

### Backend Development

```bash
cd backend

# Run with hot reload (using air)
air

# Or standard run
go run cmd/server/main.go

# Run tests
go test ./...

# Run specific test
go test -v ./internal/handlers -run TestAgentHandler

# Build binary
go build -o bin/server cmd/server/main.go

# Check for issues
go vet ./...
golangci-lint run
```

### Frontend Development

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

### E2B Service Development

```bash
cd backend/e2b-service

# Create virtual environment
python -m venv venv
source venv/bin/activate  # or `venv\Scripts\activate` on Windows

# Install dependencies
pip install -r requirements.txt

# Run service
python main.py

# Run with debug logging
DEBUG=1 python main.py
```

---

## Docker Development

### Full Stack (Recommended for Testing)

```bash
# Build and start all services
docker compose up --build

# Start in background
docker compose up -d --build

# View logs
docker compose logs -f backend
docker compose logs -f frontend

# Stop all services
docker compose down

# Reset everything (including volumes)
docker compose down -v
```

### Database Only

```bash
# Start just databases
docker compose up mongodb mysql redis -d

# Then run backend/frontend locally
```

### Rebuild Specific Service

```bash
docker compose up --build backend
docker compose up --build frontend
```

---

## Testing

### Backend Tests

```bash
cd backend

# All tests
go test ./...

# With coverage
go test -cover ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package
go test ./internal/services -v

# Integration tests (requires running databases)
go test ./internal/handlers -tags=integration
```

### Frontend Tests

```bash
cd frontend

# Unit tests
npm test

# With coverage
npm run test:coverage

# E2E tests (requires running app)
npm run test:e2e
```

### API Testing

```bash
# Health check
curl http://localhost:3001/health

# Login
curl -X POST http://localhost:3001/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@localhost","password":"admin"}'

# Authenticated request
curl http://localhost:3001/api/agents \
  -H "Authorization: Bearer <token>"
```

---

## Debugging

### Backend Debugging

**VS Code launch.json:**
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Backend",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/backend/cmd/server/main.go",
      "env": {
        "ENVIRONMENT": "development"
      }
    }
  ]
}
```

**Enable verbose logging:**
```bash
# Add to .env
LOG_LEVEL=debug
```

### Frontend Debugging

**React DevTools:**
- Install browser extension
- Use Components/Profiler tabs

**Zustand DevTools:**
```typescript
// Already enabled in development mode
// Open Redux DevTools extension
```

**Network debugging:**
- Chrome DevTools → Network tab
- Filter by WS for WebSocket messages

### Database Debugging

```bash
# MongoDB shell
docker exec -it claraverse-mongodb mongosh claraverse

# MySQL shell
docker exec -it claraverse-mysql mysql -u claraverse_user -p claraverse

# Redis CLI
docker exec -it claraverse-redis redis-cli
```

---

## Common Issues

### Port Already in Use

```bash
# Find process using port
lsof -i :3001
lsof -i :5173

# Kill process
kill -9 <PID>

# Or use different ports
PORT=3002 go run cmd/server/main.go
VITE_PORT=5174 npm run dev
```

### Database Connection Failed

```bash
# Check if containers are running
docker ps

# Check container logs
docker logs claraverse-mongodb
docker logs claraverse-mysql

# Restart containers
docker compose restart mongodb mysql redis
```

### Go Module Issues

```bash
cd backend

# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download

# Tidy up
go mod tidy
```

### Node Module Issues

```bash
cd frontend

# Clear cache
rm -rf node_modules package-lock.json
npm cache clean --force

# Reinstall
npm install
```

### WebSocket Connection Failed

1. Check CORS settings in `.env`:
   ```
   ALLOWED_ORIGINS=http://localhost:5173
   ```

2. Verify WebSocket URL in frontend:
   ```
   VITE_WS_URL=ws://localhost:3001
   ```

3. Check for rate limiting (20 connections/minute default)

### Build Failures

```bash
# Backend
cd backend
go build -v ./... 2>&1 | head -50

# Frontend
cd frontend
npm run build 2>&1 | head -50
```

---

## Code Style

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing

### TypeScript/React

- ESLint + Prettier configured
- Run `npm run lint:fix` before committing
- Use functional components with hooks

### Commit Messages

```
feat: add new feature
fix: bug fix
docs: documentation changes
style: formatting, no code change
refactor: code restructuring
test: adding tests
chore: maintenance
```

---

## Useful Commands

```bash
# Makefile shortcuts
make install      # Install all dependencies
make dev          # Start development environment
make build        # Build all services
make test         # Run all tests
make clean        # Clean build artifacts

# Database management
make db-migrate   # Run migrations
make db-reset     # Reset databases

# Docker shortcuts
make docker-up    # Start Docker services
make docker-down  # Stop Docker services
make docker-logs  # View logs
```

---

## Related Documentation

- [Architecture Guide](ARCHITECTURE.md) - System design
- [API Reference](API_REFERENCE.md) - API documentation
- [Security Guide](FINAL_SECURITY_INSPECTION.md) - Security details
- [Admin Guide](ADMIN_GUIDE.md) - Administration
- [Quick Reference](QUICK_REFERENCE.md) - Common commands
