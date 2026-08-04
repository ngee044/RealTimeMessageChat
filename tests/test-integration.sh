#!/usr/bin/env bash
#
# Integration Test Script for RealTimeMessageChat
# Requires Docker services to be running
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_OUT="$PROJECT_ROOT/build/out"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# RabbitMQ settings
RABBITMQ_HOST="${RABBITMQ_HOST:-localhost}"
RABBITMQ_PORT="${RABBITMQ_PORT:-15672}"
RABBITMQ_USER="${RABBITMQ_USER:-guest}"
RABBITMQ_PASS="${RABBITMQ_PASS:-guest}"
QUEUE_NAME="${QUEUE_NAME:-message_broadcast_queue}"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Print functions
print_header() {
    echo ""
    echo -e "${BLUE}===========================================${NC}"
    echo -e "${BLUE} $1${NC}"
    echo -e "${BLUE}===========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    ((TESTS_PASSED++)) || true
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
    ((TESTS_FAILED++)) || true
}

print_info() {
    echo -e "  $1"
}

# Cleanup function
cleanup() {
    if [ -n "${USERCLIENT_PID:-}" ] && kill -0 "$USERCLIENT_PID" 2>/dev/null; then
        kill "$USERCLIENT_PID" 2>/dev/null || true
        wait "$USERCLIENT_PID" 2>/dev/null || true
    fi
}

trap cleanup EXIT

print_header "RealTimeMessageChat Integration Test"

# Check Docker services
print_header "1. Checking Docker Services"

# Check RabbitMQ
echo -n "  RabbitMQ: "
if docker exec rabbitmq rabbitmq-diagnostics ping &>/dev/null; then
    print_success "OK"
else
    print_error "NOT RUNNING"
    echo ""
    print_error "Docker services are not running. Please start them first:"
    print_info "cd $PROJECT_ROOT/docker && ./docker-compose.sh"
    exit 1
fi

# Check Redis
echo -n "  Redis: "
if docker exec redis redis-cli ping &>/dev/null; then
    print_success "OK"
else
    print_error "NOT RUNNING"
    exit 1
fi

# Check PostgreSQL
echo -n "  PostgreSQL: "
if docker exec postgres pg_isready -U rtmc_user -d rtmc &>/dev/null; then
    print_success "OK"
else
    print_error "NOT RUNNING"
    exit 1
fi

# Check MainServer
echo -n "  MainServer (port 9876): "
if nc -z localhost 9876 &>/dev/null; then
    print_success "OK"
else
    print_error "NOT RUNNING"
    exit 1
fi

# Test Redis connection
print_header "2. Testing Redis Operations"

echo -n "  SET/GET test: "
TEST_KEY="rtmc_test_$(date +%s)"
TEST_VALUE="test_value_$(date +%s)"

SET_RESULT=$(docker exec redis redis-cli SET "$TEST_KEY" "$TEST_VALUE" 2>&1)
GET_RESULT=$(docker exec redis redis-cli GET "$TEST_KEY" 2>&1)
docker exec redis redis-cli DEL "$TEST_KEY" &>/dev/null

if [ "$GET_RESULT" = "$TEST_VALUE" ]; then
    print_success "OK"
else
    print_error "FAILED (expected: $TEST_VALUE, got: $GET_RESULT)"
fi

# Test RabbitMQ message publishing
print_header "3. Testing RabbitMQ Message Publishing"

echo -n "  Queue declaration: "
QUEUE_RESULT=$(curl -s -w "%{http_code}" -o /dev/null \
    -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
    -H "content-type: application/json" \
    -X PUT "http://$RABBITMQ_HOST:$RABBITMQ_PORT/api/queues/%2F/$QUEUE_NAME" \
    -d '{"durable": true}')

# 201: Created, 204: No Content (already exists with same config), 400: Already exists (might have different config)
if [ "$QUEUE_RESULT" = "201" ] || [ "$QUEUE_RESULT" = "204" ] || [ "$QUEUE_RESULT" = "400" ]; then
    print_success "OK"
else
    print_error "FAILED (HTTP: $QUEUE_RESULT)"
fi

# Send test message
echo -n "  Message publish: "
TEST_MESSAGE=$(cat <<EOF
{
  "id": "test_user_001",
  "sub_id": "session_$(date +%s)",
  "message": {
    "type": "test",
    "content": "Integration test message",
    "timestamp": "$(date -Iseconds)"
  }
}
EOF
)

ENCODED_MESSAGE=$(echo "$TEST_MESSAGE" | jq -c . | jq -Rs .)

PUBLISH_PAYLOAD=$(cat <<EOF
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

PUBLISH_RESULT=$(curl -s \
    -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
    -H "content-type: application/json" \
    -X POST "http://$RABBITMQ_HOST:$RABBITMQ_PORT/api/exchanges/%2F/amq.default/publish" \
    -d "$PUBLISH_PAYLOAD")

if echo "$PUBLISH_RESULT" | grep -q '"routed":true'; then
    print_success "OK"
else
    print_error "FAILED ($PUBLISH_RESULT)"
fi

print_header "4. End-to-End Broadcast"

COMPOSE_FILE="$PROJECT_ROOT/docker/docker-compose.yml"

echo -n "  UserClient container: "
docker compose -f "$COMPOSE_FILE" --profile tools up -d userclient &>/dev/null
sleep 4
if [ "$(docker inspect -f '{{.State.Running}}' userclient 2>/dev/null)" = "true" ]; then
    print_success "Running"
else
    print_error "Failed to start"
    docker logs userclient --tail 20 2>&1 | sed 's/^/      /'
fi

echo -n "  TCP connection established: "
if docker logs userclient 2>&1 | grep -q "received condition message from Server : true"; then
    print_success "OK"
else
    print_error "No connection log"
    docker logs userclient --tail 20 2>&1 | sed 's/^/      /'
fi

BASELINE_RECEIVED=$(docker logs userclient 2>&1 | grep -c "Received broadcast message" || true)
BASELINE_ROWS=$(docker exec postgres psql -U "${POSTGRES_USER:-rtmc_user}" -d "${POSTGRES_DB:-rtmc}" \
    -tAc "SELECT COUNT(*) FROM messages" 2>/dev/null || echo 0)

E2E_CONTENT="e2e_$(date +%s)"

echo -n "  POST /api/v1/messages/send: "
SEND_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/v1/messages/send \
    -H "Content-Type: application/json" \
    -d "{\"user_id\":\"e2e_user\",\"command\":\"chat_message\",\"sub_id\":\"e2e_session\",\"content\":\"$E2E_CONTENT\"}" \
    2>/dev/null || echo "000")
if [ "$SEND_CODE" = "200" ]; then
    print_success "HTTP 200"
else
    print_error "HTTP $SEND_CODE"
fi

echo -n "  Broadcast reached UserClient: "
RECEIVED=0
for _ in $(seq 1 20); do
    if docker logs userclient 2>&1 | grep -q "$E2E_CONTENT"; then
        RECEIVED=1
        break
    fi
    sleep 1
done
if [ "$RECEIVED" = "1" ]; then
    print_success "OK"
else
    print_error "Content not received within 20s"
    echo "      --- mainserver ---"; docker logs mainserver --tail 15 2>&1 | sed 's/^/      /'
    echo "      --- consumer ---";   docker logs mainserver-consumer --tail 15 2>&1 | sed 's/^/      /'
fi

echo -n "  Persisted to PostgreSQL: "
PERSISTED=0
for _ in $(seq 1 15); do
    ROWS=$(docker exec postgres psql -U "${POSTGRES_USER:-rtmc_user}" -d "${POSTGRES_DB:-rtmc}" \
        -tAc "SELECT COUNT(*) FROM messages WHERE content = '$E2E_CONTENT'" 2>/dev/null || echo 0)
    if [ "${ROWS:-0}" -ge 1 ]; then
        PERSISTED=1
        break
    fi
    sleep 1
done
if [ "$PERSISTED" = "1" ]; then
    print_success "OK"
else
    print_error "Row not found (baseline was $BASELINE_ROWS)"
fi

print_header "5. Burst Delivery (no loss)"

BURST_COUNT=20
BURST_TAG="burst_$(date +%s)"

echo -n "  Publishing $BURST_COUNT messages: "
BURST_OK=1
for i in $(seq 1 "$BURST_COUNT"); do
    CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/v1/messages/send \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"burst_user\",\"command\":\"chat_message\",\"sub_id\":\"s$i\",\"content\":\"${BURST_TAG}_$i\"}" \
        2>/dev/null || echo "000")
    [ "$CODE" = "200" ] || BURST_OK=0
done
if [ "$BURST_OK" = "1" ]; then
    print_success "OK"
else
    print_error "Some publishes failed"
fi

echo -n "  All $BURST_COUNT delivered: "
DELIVERED=0
for _ in $(seq 1 30); do
    DELIVERED=$(docker logs userclient 2>&1 | grep -c "$BURST_TAG" || true)
    [ "${DELIVERED:-0}" -ge "$BURST_COUNT" ] && break
    sleep 1
done
if [ "${DELIVERED:-0}" -ge "$BURST_COUNT" ]; then
    print_success "$DELIVERED/$BURST_COUNT"
else
    print_error "$DELIVERED/$BURST_COUNT (loss detected)"
fi

echo -n "  Redis queue drained: "
QUEUE_LEN=$(docker exec redis redis-cli LLEN send_global_message 2>/dev/null | tr -d '\r' || echo "?")
if [ "$QUEUE_LEN" = "0" ]; then
    print_success "LLEN=0"
else
    print_error "LLEN=$QUEUE_LEN"
fi

print_header "6. Queue Topology"

echo -n "  Queue declared with auto_delete=false: "
QUEUE_INFO=$(curl -s -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
    "http://$RABBITMQ_HOST:$RABBITMQ_PORT/api/queues/%2F/$QUEUE_NAME" 2>/dev/null || echo "")
if echo "$QUEUE_INFO" | grep -q '"auto_delete":false' && echo "$QUEUE_INFO" | grep -q '"durable":true'; then
    print_success "OK"
else
    print_error "Unexpected queue arguments: $(echo "$QUEUE_INFO" | head -c 200)"
fi

echo -n "  No 406 PRECONDITION_FAILED in logs: "
if docker logs mainserver-consumer 2>&1 | grep -qi "PRECONDITION_FAILED"; then
    print_error "Found channel declaration conflict"
else
    print_success "OK"
fi

print_header "7. Health and Logs"

echo -n "  API health: "
HEALTH_RESULT=$(curl -s http://localhost:8080/health 2>/dev/null || echo "")
if echo "$HEALTH_RESULT" | grep -q '"status":"healthy"'; then
    print_success "healthy"
else
    print_error "Unexpected: $(echo "$HEALTH_RESULT" | head -c 200)"
fi

for service in mainserver mainserver-consumer; do
    echo -n "  $service logs: "
    if docker logs "$service" --tail 5 &>/dev/null; then
        print_success "Accessible"
    else
        print_error "Not accessible"
    fi
done

echo -n "  No crash markers: "
if docker logs mainserver 2>&1 | grep -qE "terminate|SIGSEGV|future_error"; then
    print_error "Found crash marker in mainserver"
elif docker logs mainserver-consumer 2>&1 | grep -qE "terminate|SIGSEGV|future_error"; then
    print_error "Found crash marker in consumer"
else
    print_success "OK"
fi

# Summary
print_header "Test Summary"
echo ""
echo -e "  ${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "  ${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    print_success "All tests passed!"
    exit 0
else
    print_error "Some tests failed"
    exit 1
fi
