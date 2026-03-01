#!/bin/bash
# ============================================================================
# ClaraVerse Docker Development Environment
# ============================================================================
# This script starts the development environment with HMR for both
# frontend (Vite) and backend (Go with Air).
#
# Usage:
#   ./dev-docker.sh         # Start development environment
#   ./dev-docker.sh build   # Rebuild and start
#   ./dev-docker.sh down    # Stop all containers
#   ./dev-docker.sh clean   # Remove all resources except MySQL data
#   ./dev-docker.sh logs    # View logs
#   ./dev-docker.sh logs -f # Follow logs
# ============================================================================

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check for .env file
if [ ! -f .env ]; then
    echo -e "${RED}❌ Error: .env file not found!${NC}"
    echo ""
    echo -e "${YELLOW}Please create a .env file with the required configuration.${NC}"
    echo -e "${YELLOW}You can use .env.example as a template:${NC}"
    echo ""
    echo -e "${CYAN}  cp .env.example .env${NC}"
    echo ""
    echo -e "${YELLOW}Then edit .env and add your configuration values.${NC}"
    echo ""
    exit 1
fi

# Use 'docker compose' (V2) or 'docker-compose' (V1)
if command -v docker &> /dev/null && docker compose version &> /dev/null; then
    COMPOSE="docker compose"
elif command -v docker-compose &> /dev/null; then
    COMPOSE="docker-compose"
else
    echo "Error: Neither 'docker compose' nor 'docker-compose' found."
    echo "Please install Docker with Compose plugin."
    exit 1
fi

case "${1:-up}" in
  up)
    echo -e "${CYAN}🚀 Starting ClaraVerse Development Environment...${NC}"
    echo -e "${YELLOW}   Frontend: http://localhost:5173 (Vite HMR)${NC}"
    echo -e "${YELLOW}   Backend:  http://localhost:3001 (Go Air)${NC}"
    echo ""
    $COMPOSE -f docker-compose.yml -f docker-compose.dev.yml up
    ;;
  build)
    echo -e "${CYAN}🔨 Rebuilding and starting development environment...${NC}"
    $COMPOSE -f docker-compose.yml -f docker-compose.dev.yml up --build
    ;;
  down)
    echo -e "${CYAN}🛑 Stopping development environment...${NC}"
    $COMPOSE -f docker-compose.yml -f docker-compose.dev.yml down
    ;;
  clean)
    echo -e "${CYAN}🧹 Cleaning up Docker resources (preserving MySQL data)...${NC}"
    echo -e "${YELLOW}⚠️  This will remove:${NC}"
    echo -e "${YELLOW}   - All containers${NC}"
    echo -e "${YELLOW}   - All volumes EXCEPT mysql-data-new${NC}"
    echo -e "${YELLOW}   - Network${NC}"
    echo -e "${YELLOW}   - Built images${NC}"
    echo ""
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      # Stop and remove containers
      $COMPOSE -f docker-compose.yml -f docker-compose.dev.yml down

      # Remove specific volumes (all except mysql-data-new)
      echo -e "${CYAN}Removing volumes...${NC}"
      docker volume rm claraverse-scarlet_backend-data 2>/dev/null || true
      docker volume rm claraverse-scarlet_backend-uploads 2>/dev/null || true
      docker volume rm claraverse-scarlet_backend-logs 2>/dev/null || true
      docker volume rm claraverse-scarlet_mongodb-data 2>/dev/null || true
      docker volume rm claraverse-scarlet_redis-data 2>/dev/null || true

      # Remove old mysql-data volume if it exists
      echo -e "${CYAN}Removing old MySQL volume...${NC}"
      docker volume rm claraverse-scarlet_mysql-data 2>/dev/null || true

      # Force remove network
      echo -e "${CYAN}Removing network...${NC}"
      docker network rm claraverse-scarlet_claraverse-network 2>/dev/null || true

      # Remove built images
      echo -e "${CYAN}Removing built images...${NC}"
      docker rmi claraverse-scarlet-backend 2>/dev/null || true
      docker rmi claraverse-scarlet-frontend 2>/dev/null || true
      docker rmi claraverse-scarlet-e2b-service 2>/dev/null || true

      echo -e "${GREEN}✅ Cleanup complete! MySQL data preserved in mysql-data-new volume.${NC}"
    else
      echo -e "${YELLOW}Cleanup cancelled.${NC}"
    fi
    ;;
  logs)
    shift
    $COMPOSE -f docker-compose.yml -f docker-compose.dev.yml logs "$@"
    ;;
  restart)
    echo -e "${CYAN}🔄 Restarting development environment...${NC}"
    $COMPOSE -f docker-compose.yml -f docker-compose.dev.yml restart
    ;;
  *)
    echo "Usage: $0 {up|build|down|clean|logs|restart}"
    echo ""
    echo "Commands:"
    echo "  up      - Start development environment (default)"
    echo "  build   - Rebuild containers and start"
    echo "  down    - Stop all containers"
    echo "  clean   - Remove all resources except MySQL data"
    echo "  logs    - View container logs (add -f to follow)"
    echo "  restart - Restart all containers"
    exit 1
    ;;
esac

