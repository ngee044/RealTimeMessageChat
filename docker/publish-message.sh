#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# RabbitMQ settings
RABBITMQ_HOST="${RABBITMQ_HOST:-localhost}"
RABBITMQ_PORT="${RABBITMQ_PORT:-15672}"
RABBITMQ_USER="${RABBITMQ_USER:-guest}"
RABBITMQ_PASS="${RABBITMQ_PASS:-guest}"
QUEUE_NAME="${QUEUE_NAME:-message_broadcast_queue}"

MESSAGE_CONTENT="${1:-Hello from RealTimeMessageChat test!}"
USER_ID="${2:-test_user}"
COMMAND="${3:-chat_message}"

echo "=========================================="
echo " Publish Message to RabbitMQ"
echo "=========================================="
echo ""
echo "Queue: $QUEUE_NAME"
echo "User ID: $USER_ID"
echo "Command: $COMMAND"
echo "Content: $MESSAGE_CONTENT"
echo ""

# Create queue if not exists and publish message using RabbitMQ Management API
# First, declare the queue
curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
    -H "content-type: application/json" \
    -X PUT "http://$RABBITMQ_HOST:$RABBITMQ_PORT/api/queues/%2F/$QUEUE_NAME" \
    -d '{"durable": true}' > /dev/null

# Build the message in the format expected by MainServerConsumer
MESSAGE_JSON=$(cat <<EOF
{
  "id": "$USER_ID",
  "sub_id": "session_$(date +%s)",
  "message": {
    "command": "$COMMAND",
    "content": "$MESSAGE_CONTENT",
    "timestamp": "$(date -Iseconds)"
  }
}
EOF
)

# Escape the JSON for embedding in the payload
ENCODED_MESSAGE=$(echo "$MESSAGE_JSON" | jq -c . | jq -Rs .)

# Publish message
PAYLOAD=$(cat <<EOF
{
  "properties": {
    "delivery_mode": 2,
    "content_type": "application/json"
  },
  "routing_key": "$QUEUE_NAME",
  "payload": $ENCODED_MESSAGE,
  "payload_encoding": "string"
}
EOF
)

RESULT=$(curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
    -H "content-type: application/json" \
    -X POST "http://$RABBITMQ_HOST:$RABBITMQ_PORT/api/exchanges/%2F/amq.default/publish" \
    -d "$PAYLOAD")

if echo "$RESULT" | grep -q '"routed":true'; then
    echo "Message published successfully!"
else
    echo "Failed to publish message:"
    echo "$RESULT"
    exit 1
fi

echo ""
