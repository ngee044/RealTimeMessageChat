#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_OUT="$PROJECT_ROOT/build/out"
CLIENT_BIN="$BUILD_OUT/UserClient"
CONFIG_FILE="$SCRIPT_DIR/config/user_client_configurations.json"

echo "=========================================="
echo " RealTimeMessageChat UserClient"
echo "=========================================="
echo ""

# Check if UserClient binary exists
if [ ! -f "$CLIENT_BIN" ]; then
    echo "ERROR: UserClient binary not found at $CLIENT_BIN"
    echo ""
    echo "Please build the project first:"
    echo "  cd $PROJECT_ROOT"
    echo "  ./build.sh"
    echo "  cd build && cmake --build . --config Release"
    echo ""
    exit 1
fi

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "ERROR: Configuration file not found at $CONFIG_FILE"
    exit 1
fi

# Check if MainServer is running
echo "Checking MainServer connection..."
if ! nc -z localhost 9876 &>/dev/null; then
    echo "WARNING: MainServer does not appear to be running on localhost:9876"
    echo "Please start Docker services first: ./docker-compose.sh"
    echo ""
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo "Starting UserClient..."
echo "Config: $CONFIG_FILE"
echo ""
echo "Press Ctrl+C to stop"
echo "=========================================="
echo ""

# Copy config to build directory and run
cp "$CONFIG_FILE" "$BUILD_OUT/user_client_configurations.json"
cd "$BUILD_OUT"
./UserClient
