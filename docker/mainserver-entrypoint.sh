#!/bin/bash
set -e

# Update configuration file with environment variables
CONFIG_FILE="/app/main_server_configurations.json"

if [ -f "$CONFIG_FILE" ]; then
    # Use environment variables or defaults
    REDIS_HOST=${REDIS_HOST:-redis}
    REDIS_PORT=${REDIS_PORT:-6379}
    RABBITMQ_HOST=${RABBITMQ_HOST:-rabbitmq}
    RABBITMQ_PORT=${RABBITMQ_PORT:-5672}
    SERVER_PORT=${SERVER_PORT:-9876}

    # Update Redis configuration
    sed -i "s/\"redis_host\": \".*\"/\"redis_host\": \"$REDIS_HOST\"/" "$CONFIG_FILE"
    sed -i "s/\"redis_port\": .*/\"redis_port\": $REDIS_PORT,/" "$CONFIG_FILE"

    echo "MainServer configuration updated:"
    echo "  Redis: $REDIS_HOST:$REDIS_PORT"
    echo "  Server Port: $SERVER_PORT"
fi

# Execute the main command
exec "$@"
