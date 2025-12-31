#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

SERVICE="${1:-all}"

case "$SERVICE" in
    mainserver)
        echo "=== MainServer Logs ==="
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" logs -f mainserver
        ;;
    consumer)
        echo "=== MainServerConsumer Logs ==="
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" logs -f mainserver-consumer
        ;;
    api-server)
        echo "=== API Server Logs ==="
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" logs -f api-server
        ;;
    rabbitmq)
        echo "=== RabbitMQ Logs ==="
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" logs -f rabbitmq
        ;;
    redis)
        echo "=== Redis Logs ==="
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" logs -f redis
        ;;
    postgres)
        echo "=== PostgreSQL Logs ==="
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" logs -f postgres
        ;;
    all)
        echo "=== All Service Logs ==="
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" logs -f
        ;;
    *)
        echo "Usage: $0 [mainserver|consumer|api-server|rabbitmq|redis|postgres|all]"
        exit 1
        ;;
esac
