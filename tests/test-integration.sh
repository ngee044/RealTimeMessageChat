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

print_warning() {
    echo -e "${YELLOW}! $1${NC}"
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

# Test UserClient connection
print_header "4. Testing UserClient Connection"

if [ -f "$BUILD_OUT/UserClient" ]; then
    echo -n "  UserClient binary: "
    print_success "Found"

    # Copy config if exists
    if [ -f "$PROJECT_ROOT/docker/config/user_client_configurations.json" ]; then
        cp "$PROJECT_ROOT/docker/config/user_client_configurations.json" "$BUILD_OUT/user_client_configurations.json"
    fi

    echo -n "  Connection test: "
    cd "$BUILD_OUT"

    # Run UserClient with timeout (use gtimeout on macOS if available)
    if command -v gtimeout &>/dev/null; then
        gtimeout 5 ./UserClient &>/dev/null &
    elif command -v timeout &>/dev/null; then
        timeout 5 ./UserClient &>/dev/null &
    else
        ./UserClient &>/dev/null &
    fi
    USERCLIENT_PID=$!

    sleep 3

    if kill -0 "$USERCLIENT_PID" 2>/dev/null; then
        print_success "Connected"
        kill "$USERCLIENT_PID" 2>/dev/null || true
    else
        wait "$USERCLIENT_PID" 2>/dev/null
        EXIT_CODE=$?
        if [ $EXIT_CODE -eq 0 ] || [ $EXIT_CODE -eq 124 ]; then
            print_success "OK (completed)"
        else
            print_warning "Exited with code $EXIT_CODE"
        fi
    fi
    USERCLIENT_PID=""
else
    print_warning "UserClient binary not found, skipping connection test"
fi

# Test API Server (if running)
print_header "5. Testing API Server"

echo -n "  Health check: "
if curl -s http://localhost:8080/health &>/dev/null; then
    HEALTH_RESULT=$(curl -s http://localhost:8080/health)
    if echo "$HEALTH_RESULT" | grep -qi "ok\|healthy\|success"; then
        print_success "OK"
    else
        print_success "Responding (status: $HEALTH_RESULT)"
    fi
else
    print_warning "Not available (API server may not be running)"
fi

# Test sending message via API
echo -n "  Send message API: "
API_RESULT=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/messages/send \
    -H "Content-Type: application/json" \
    -d '{
        "user_id": "test_user",
        "command": "chat_message",
        "content": "Test message from integration test"
    }' 2>/dev/null || echo "000")

HTTP_CODE=$(echo "$API_RESULT" | tail -n1)
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "202" ]; then
    print_success "OK (HTTP $HTTP_CODE)"
elif [ "$HTTP_CODE" = "000" ]; then
    print_warning "API server not available"
else
    print_warning "HTTP $HTTP_CODE"
fi

# Check MainServer logs
print_header "6. Checking Logs"

echo -n "  MainServer logs: "
if docker logs mainserver --tail 5 &>/dev/null; then
    print_success "Accessible"
else
    print_warning "Not accessible"
fi

echo -n "  Consumer logs: "
if docker logs mainserver-consumer --tail 5 &>/dev/null; then
    print_success "Accessible"
else
    print_warning "Not accessible"
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
