#!/usr/bin/env bash

# RealTimeMessageChat Docker Clean Script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_ARGS=(-f "$SCRIPT_DIR/docker-compose.yml")

echo "======================================"
echo "RealTimeMessageChat Docker Clean"
echo "======================================"

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

compose() {
    if docker compose version >/dev/null 2>&1; then
        docker compose "${COMPOSE_ARGS[@]}" "$@"
    else
        docker-compose "${COMPOSE_ARGS[@]}" "$@"
    fi
}

echo -e "${YELLOW}Stopping and removing all containers, networks, and volumes...${NC}"
compose down -v

echo -e "${YELLOW}Removing built images...${NC}"
compose down --rmi local

echo -e "${RED}✓ Cleanup completed!${NC}"
echo ""
echo "To rebuild and start, run:"
echo "  ./docker/build.sh"
echo "  ./docker/start.sh"
