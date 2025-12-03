#!/bin/bash
set -e

# Update configuration file with environment variables
CONFIG_FILE="/app/main_server_consumer_configurations.json"

if [ -f "$CONFIG_FILE" ]; then
    # Use environment variables or defaults
    REDIS_HOST=${REDIS_HOST:-redis}
    REDIS_PORT=${REDIS_PORT:-6379}
    RABBITMQ_HOST=${RABBITMQ_HOST:-rabbitmq}
    RABBITMQ_PORT=${RABBITMQ_PORT:-5672}
    RABBITMQ_USER=${RABBITMQ_USER:-guest}
    RABBITMQ_PASSWORD=${RABBITMQ_PASSWORD:-guest}

    # Update Redis configuration
    sed -i "s/\"redis_host\": \".*\"/\"redis_host\": \"$REDIS_HOST\"/" "$CONFIG_FILE"
    sed -i "s/\"redis_port\": .*/\"redis_port\": $REDIS_PORT,/" "$CONFIG_FILE"

    # Update RabbitMQ configuration
    sed -i "s/\"rabbit_mq_host\": \".*\"/\"rabbit_mq_host\": \"$RABBITMQ_HOST\"/" "$CONFIG_FILE"
    sed -i "s/\"rabbit_mq_port\": .*/\"rabbit_mq_port\": $RABBITMQ_PORT,/" "$CONFIG_FILE"
    sed -i "s/\"rabbit_mq_user_name\": \".*\"/\"rabbit_mq_user_name\": \"$RABBITMQ_USER\"/" "$CONFIG_FILE"
    sed -i "s/\"rabbit_mq_password\": \".*\"/\"rabbit_mq_password\": \"$RABBITMQ_PASSWORD\"/" "$CONFIG_FILE"

    echo "MainServerConsumer configuration updated:"
    echo "  Redis: $REDIS_HOST:$REDIS_PORT"
    echo "  RabbitMQ: $RABBITMQ_HOST:$RABBITMQ_PORT"
fi

# Execute the main command
exec "$@"
