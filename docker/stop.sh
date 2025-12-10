#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Stopping all RealTimeMessageChat services..."
docker compose -f "$SCRIPT_DIR/docker-compose.yml" down

echo ""
echo "All services stopped."
