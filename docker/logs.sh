#!/usr/bin/env bash

# RealTimeMessageChat Docker Logs Script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_ARGS=(-f "$SCRIPT_DIR/docker-compose.yml")
SERVICE=${1:-""}

compose() {
    if docker compose version >/dev/null 2>&1; then
        docker compose "${COMPOSE_ARGS[@]}" "$@"
    else
        docker-compose "${COMPOSE_ARGS[@]}" "$@"
    fi
}

if [ -z "$SERVICE" ]; then
    echo "Showing logs for all services (Ctrl+C to exit)..."
    compose logs -f
else
    echo "Showing logs for $SERVICE (Ctrl+C to exit)..."
    compose logs -f "$SERVICE"
fi
