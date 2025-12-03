#!/usr/bin/env bash

# RealTimeMessageChat Docker Start Script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_ARGS=(-f "$SCRIPT_DIR/docker-compose.yml")

echo "======================================"
echo "RealTimeMessageChat Docker Start"
echo "======================================"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

compose() {
    if docker compose version >/dev/null 2>&1; then
        docker compose "${COMPOSE_ARGS[@]}" "$@"
    else
        docker-compose "${COMPOSE_ARGS[@]}" "$@"
    fi
}

# Start services
echo -e "${YELLOW}Starting all services...${NC}"
compose up -d

echo ""
echo -e "${GREEN}✓ All services started!${NC}"
echo ""
echo -e "${BLUE}Service Status:${NC}"
compose ps

echo ""
echo -e "${BLUE}Service URLs:${NC}"
echo "  MainServer (TCP):        localhost:9876"
echo "  Redis:                   localhost:6379"
echo "  RabbitMQ:                localhost:5672"
echo "  RabbitMQ Management UI:  http://localhost:15672 (guest/guest)"
echo ""
echo "To view logs:"
echo "  docker compose -f \"$SCRIPT_DIR/docker-compose.yml\" logs -f mainserver     # MainServer logs"
echo "  docker compose -f \"$SCRIPT_DIR/docker-compose.yml\" logs -f mainserver-consumer # Consumer logs"
echo "  docker compose -f \"$SCRIPT_DIR/docker-compose.yml\" logs -f                # All logs"
echo ""
echo "To stop services:"
echo "  ./docker/stop.sh"
