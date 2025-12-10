#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Create log directories
mkdir -p "$SCRIPT_DIR/logs/mainserver"
mkdir -p "$SCRIPT_DIR/logs/consumer"

echo "=========================================="
echo " RealTimeMessageChat Docker Deployment"
echo "=========================================="
echo ""
echo "Building and starting services..."
echo ""

docker compose -f "$SCRIPT_DIR/docker-compose.yml" up --build -d

echo ""
echo "Waiting for services to be ready..."
sleep 5

echo ""
echo "Service Status:"
echo "---------------"
docker compose -f "$SCRIPT_DIR/docker-compose.yml" ps

echo ""
echo "=========================================="
echo " Deployment Complete!"
echo "=========================================="
echo ""
echo "Services running:"
echo "  - RabbitMQ Management: http://localhost:15672 (guest/guest)"
echo "  - Redis: localhost:6379"
echo "  - MainServer: localhost:9876"
echo ""
echo "To run UserClient locally:"
echo "  ./run-client.sh"
echo ""
echo "To view logs:"
echo "  ./logs.sh [mainserver|consumer|rabbitmq|redis|all]"
echo ""
echo "To stop all services:"
echo "  ./stop.sh"
echo ""
