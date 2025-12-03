#!/usr/bin/env bash

# RealTimeMessageChat Docker Stop Script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_ARGS=(-f "$SCRIPT_DIR/docker-compose.yml")

echo "======================================"
echo "RealTimeMessageChat Docker Stop"
echo "======================================"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

compose() {
    if docker compose version >/dev/null 2>&1; then
        docker compose "${COMPOSE_ARGS[@]}" "$@"
    else
        docker-compose "${COMPOSE_ARGS[@]}" "$@"
    fi
}

echo -e "${YELLOW}Stopping all services...${NC}"
compose down

echo -e "${GREEN}✓ All services stopped!${NC}"
