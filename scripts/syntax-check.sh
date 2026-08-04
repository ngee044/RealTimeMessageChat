#!/usr/bin/env bash

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT"

TRIPLET="${VCPKG_TRIPLET:-}"
if [[ -z "$TRIPLET" ]]; then
    for candidate in "$ROOT"/build/vcpkg_installed/*/include; do
        [[ -d "$candidate" ]] || continue
        TRIPLET="$(basename "$(dirname "$candidate")")"
        break
    done
fi

VCPKG_INC="$ROOT/build/vcpkg_installed/$TRIPLET/include"
if [[ ! -d "$VCPKG_INC" ]]; then
    echo "error: vcpkg include dir not found: $VCPKG_INC" >&2
    echo "       run ./build.sh (configure 단계까지만 성공해도 충분하다)" >&2
    exit 1
fi

CXX="${CXX_SYNTAX:-}"
if [[ -z "$CXX" ]]; then
    for candidate in /opt/homebrew/opt/llvm/bin/clang++ /usr/local/opt/llvm/bin/clang++; do
        if [[ -x "$candidate" ]]; then CXX="$candidate"; break; fi
    done
fi
if [[ -z "$CXX" ]]; then
    CXX="$(command -v clang++ || true)"
    echo "warning: LLVM clang++ 을 찾지 못해 $CXX 로 대체한다 (Apple clang 이면 실패할 수 있다)" >&2
fi
if [[ -z "$CXX" ]]; then
    echo "error: no C++ compiler found" >&2
    exit 1
fi

TK="$ROOT/.cpp_tool_kit"

FLAGS=(
    -std=c++23
    -fsyntax-only
    -DUSE_ENCRYPT_MODULE
    -DCRYPTOPP_INCLUDE_PREFIX=cryptopp
    -DBOOST_JSON_NO_LIB=1
    -DBOOST_JSON_STATIC_LINK=1
    -DBOOST_FILESYSTEM_NO_LIB
    -DBOOST_CONTAINER_NO_LIB
)
INCS=(
    -I"$TK/Utilities"
    -I"$TK/ThreadPool"
    -I"$TK/Network"
    -I"$TK/Database"
    -I"$TK/RabbitMQ"
    -I"$TK/Redis"
    -I"$ROOT/CommonModule"
    -isystem "$VCPKG_INC"
)

collect_sources() {
    local dir="$1" cmake="$1/CMakeLists.txt"
    [[ -f "$cmake" ]] || return 0
    awk '
        /SOURCE_FILES/ { inside = 1 }
        inside {
            if (match($0, /[A-Za-z0-9_]+\.cpp/)) {
                print substr($0, RSTART, RLENGTH)
            }
            if (/\)/) { inside = 0 }
        }
    ' "$cmake" | while read -r src; do
        [[ -f "$dir/$src" ]] && echo "$dir/$src"
    done
}

SOURCES=()
for component in CommonModule MainServer MainServerConsumer UserClient; do
    while IFS= read -r src; do
        [[ -n "$src" ]] && SOURCES+=("$src")
    done < <(collect_sources "$component")
done

if [[ ${#SOURCES[@]} -eq 0 ]]; then
    echo "error: no sources collected from CMakeLists.txt" >&2
    exit 1
fi

echo "compiler : $CXX"
echo "vcpkg    : $VCPKG_INC"
echo "targets  : ${#SOURCES[@]} translation units"
echo

pass=0
fail=0
failed_units=()
log="$(mktemp)"
trap 'rm -f "$log"' EXIT

for src in "${SOURCES[@]}"; do
    printf '%-52s ' "$src"
    if "$CXX" "${FLAGS[@]}" "${INCS[@]}" -I"$(dirname "$src")" "$src" > "$log" 2>&1; then
        echo "PASS"
        pass=$((pass + 1))
    else
        echo "FAIL"
        fail=$((fail + 1))
        failed_units+=("$src")
        grep -E 'error:' "$log" | head -5 | sed 's/^/    /'
    fi
done

echo
echo "result: $pass passed, $fail failed"

if [[ $fail -gt 0 ]]; then
    echo "failed units:"
    printf '  %s\n' "${failed_units[@]}"
    exit 1
fi
