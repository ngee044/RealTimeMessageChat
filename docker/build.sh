#!/usr/bin/env bash

# RealTimeMessageChat Docker Build Script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_ARGS=(-f "$SCRIPT_DIR/docker-compose.yml")

echo "======================================"
echo "RealTimeMessageChat Docker Build"
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

# Build the shared image (used by both services)
echo -e "${YELLOW}Building Docker image...${NC}"
compose build mainserver

echo -e "${GREEN}✓ Build completed successfully!${NC}"
echo ""
echo "To start the services, run:"
echo "  ./docker/start.sh"
