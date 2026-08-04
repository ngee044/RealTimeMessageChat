#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="${BUILD_DIR:-$SCRIPT_DIR/build}"
BUILD_TYPE="${BUILD_TYPE:-Release}"

# 툴체인 경로는 cwd 에 의존하지 않도록 절대경로로 해석한다.
VCPKG_ROOT="${VCPKG_ROOT:-$HOME/vcpkg}"
TOOLCHAIN="$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake"

if [[ ! -f "$TOOLCHAIN" ]]; then
    echo "error: vcpkg toolchain not found at $TOOLCHAIN" >&2
    echo "       set VCPKG_ROOT to your vcpkg checkout" >&2
    exit 1
fi

GENERATOR_ARGS=()
if command -v ninja >/dev/null 2>&1; then
    GENERATOR_ARGS=(-G Ninja)
fi

rm -rf "$BUILD_DIR"

cmake -S "$SCRIPT_DIR" -B "$BUILD_DIR" \
    "${GENERATOR_ARGS[@]}" \
    -DCMAKE_TOOLCHAIN_FILE="$TOOLCHAIN" \
    -DCMAKE_BUILD_TYPE="$BUILD_TYPE" \
    -DBUILD_SHARED_LIBS=OFF

cmake --build "$BUILD_DIR" --config "$BUILD_TYPE" --parallel

echo ""
echo "Build complete. Binaries: $BUILD_DIR/out"
