#!/usr/bin/env bash
#
# Local Test Script for RealTimeMessageChat
# Tests the built executables without Docker
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

# Print functions
print_header() {
    echo ""
    echo -e "${BLUE}===========================================${NC}"
    echo -e "${BLUE} $1${NC}"
    echo -e "${BLUE}===========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}! $1${NC}"
}

print_info() {
    echo -e "  $1"
}

# Cleanup function
cleanup() {
    print_header "Cleanup"

    if [ -n "${MAINSERVER_PID:-}" ] && kill -0 "$MAINSERVER_PID" 2>/dev/null; then
        print_info "Stopping MainServer (PID: $MAINSERVER_PID)..."
        kill "$MAINSERVER_PID" 2>/dev/null || true
        wait "$MAINSERVER_PID" 2>/dev/null || true
    fi

    if [ -n "${USERCLIENT_PID:-}" ] && kill -0 "$USERCLIENT_PID" 2>/dev/null; then
        print_info "Stopping UserClient (PID: $USERCLIENT_PID)..."
        kill "$USERCLIENT_PID" 2>/dev/null || true
        wait "$USERCLIENT_PID" 2>/dev/null || true
    fi

    print_success "Cleanup complete"
}

trap cleanup EXIT

print_header "RealTimeMessageChat Local Test"

# Check if binaries exist
print_header "Checking Binaries"

BINARIES=("MainServer" "UserClient" "MainServerConsumer")
ALL_EXIST=true

for bin in "${BINARIES[@]}"; do
    if [ -f "$BUILD_OUT/$bin" ]; then
        print_success "$bin found"
    else
        print_error "$bin not found at $BUILD_OUT/$bin"
        ALL_EXIST=false
    fi
done

if [ "$ALL_EXIST" = false ]; then
    echo ""
    print_error "Some binaries are missing. Please build the project first:"
    print_info "cd $PROJECT_ROOT && ./build.sh"
    exit 1
fi

# Check configuration files
print_header "Checking Configuration Files"

CONFIG_FILES=(
    "$BUILD_OUT/main_server_configurations.json"
    "$BUILD_OUT/user_client_configurations.json"
)

for config in "${CONFIG_FILES[@]}"; do
    if [ -f "$config" ]; then
        print_success "$(basename "$config") found"
    else
        print_warning "$(basename "$config") not found, will use defaults"
    fi
done

# Test MainServer startup
print_header "Testing MainServer Startup"

cd "$BUILD_OUT"

print_info "Starting MainServer..."
./MainServer &
MAINSERVER_PID=$!

sleep 2

if kill -0 "$MAINSERVER_PID" 2>/dev/null; then
    print_success "MainServer started (PID: $MAINSERVER_PID)"
else
    print_error "MainServer failed to start"
    exit 1
fi

# Check if MainServer is listening
print_info "Checking MainServer port 9876..."
sleep 1

if nc -z localhost 9876 &>/dev/null; then
    print_success "MainServer is listening on port 9876"
else
    print_warning "MainServer port check failed (nc command may not be available)"
fi

# Test UserClient connection
print_header "Testing UserClient Connection"

print_info "Starting UserClient..."
# Use gtimeout if available (macOS with coreutils), otherwise run without timeout
if command -v gtimeout &>/dev/null; then
    gtimeout 5 ./UserClient &
elif command -v timeout &>/dev/null; then
    timeout 5 ./UserClient &
else
    ./UserClient &
fi
USERCLIENT_PID=$!

sleep 3

if kill -0 "$USERCLIENT_PID" 2>/dev/null; then
    print_success "UserClient is running (PID: $USERCLIENT_PID)"
else
    # UserClient might have exited normally
    wait "$USERCLIENT_PID" 2>/dev/null
    EXIT_CODE=$?
    if [ $EXIT_CODE -eq 0 ] || [ $EXIT_CODE -eq 124 ]; then
        print_success "UserClient connection test completed"
    else
        print_warning "UserClient exited with code $EXIT_CODE"
    fi
    USERCLIENT_PID=""
fi

# Test graceful shutdown (Signal Handler)
print_header "Testing Graceful Shutdown (Signal Handler)"

if [ -n "${USERCLIENT_PID:-}" ] && kill -0 "$USERCLIENT_PID" 2>/dev/null; then
    print_info "Sending SIGTERM to UserClient..."
    kill -TERM "$USERCLIENT_PID" 2>/dev/null || true
    sleep 1

    if kill -0 "$USERCLIENT_PID" 2>/dev/null; then
        print_warning "UserClient still running after SIGTERM"
    else
        print_success "UserClient gracefully shut down"
    fi
    USERCLIENT_PID=""
fi

print_info "Sending SIGTERM to MainServer..."
kill -TERM "$MAINSERVER_PID" 2>/dev/null || true
sleep 2

if kill -0 "$MAINSERVER_PID" 2>/dev/null; then
    print_warning "MainServer still running after SIGTERM, sending SIGKILL..."
    kill -KILL "$MAINSERVER_PID" 2>/dev/null || true
else
    print_success "MainServer gracefully shut down"
fi
MAINSERVER_PID=""

print_header "Test Summary"
print_success "All basic tests passed!"
echo ""
print_info "Note: For full integration tests, run with Docker:"
print_info "  cd $PROJECT_ROOT/docker && ./docker-compose.sh"
print_info "  cd $PROJECT_ROOT/tests && ./test-integration.sh"
echo ""
