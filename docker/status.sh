#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=========================================="
echo " RealTimeMessageChat Service Status"
echo "=========================================="
echo ""

docker compose -f "$SCRIPT_DIR/docker-compose.yml" ps

echo ""
echo "=========================================="
echo " Health Check"
echo "=========================================="

# Check RabbitMQ
echo -n "RabbitMQ: "
if docker exec rabbitmq rabbitmq-diagnostics ping &>/dev/null; then
    echo "OK"
else
    echo "NOT READY"
fi

# Check Redis
echo -n "Redis: "
if docker exec redis redis-cli ping &>/dev/null; then
    echo "OK"
else
    echo "NOT READY"
fi

# Check MainServer port
echo -n "MainServer (port 9876): "
if nc -z localhost 9876 &>/dev/null; then
    echo "OK"
else
    echo "NOT READY"
fi

echo ""
