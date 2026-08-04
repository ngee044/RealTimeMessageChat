#!/usr/bin/env bash
#
# rtmc_test.sh — 로컬 빌드 산출물로 브로드캐스트 시연/검증
#
#   build/out 의 MainServer / MainServerConsumer / UserClient 를 호스트에서 직접 실행하고,
#   클라이언트 3개가 RabbitMQ -> Consumer -> Redis -> MainServer -> TCP 경로로 온
#   동일 메시지를 각각 수신했는지 로그로 확인한다.
#
#   Redis / RabbitMQ 는 docker compose 로 띄운다 (PostgreSQL 은 use_database=false 라 불필요).
#
# 사용법:
#   ./rtmc_test.sh                 클라이언트 3개 / 메시지 5건
#   CLIENTS=5 MESSAGES=20 ./rtmc_test.sh
#   KEEP_RUNNING=1 ./rtmc_test.sh  검증 후 프로세스를 남겨 둔다 (수동 시연용)

set -uo pipefail

CLIENTS="${CLIENTS:-3}"
MESSAGES="${MESSAGES:-5}"
KEEP_RUNNING="${KEEP_RUNNING:-0}"
DELIVER_TIMEOUT="${DELIVER_TIMEOUT:-30}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_OUT="$SCRIPT_DIR/build/out"
RUN_DIR="$SCRIPT_DIR/build/rtmc_test"
LOG_DIR="$RUN_DIR/logs"
COMPOSE="$SCRIPT_DIR/docker/docker-compose.yml"
TODAY="$(date +%Y-%m-%d)"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
PASSED=0; FAILED=0
PIDS=()

header() { echo; echo -e "${BLUE}==========================================${NC}"; echo -e "${BLUE} $1${NC}"; echo -e "${BLUE}==========================================${NC}"; }
ok()     { echo -e "  ${GREEN}✓${NC} $1"; PASSED=$((PASSED + 1)); }
fail()   { echo -e "  ${RED}✗${NC} $1"; FAILED=$((FAILED + 1)); }
warn()   { echo -e "  ${YELLOW}!${NC} $1"; }
info()   { echo "    $1"; }

cleanup() {
    if [ "$KEEP_RUNNING" = "1" ] && [ "$FAILED" -eq 0 ]; then
        return
    fi
    for pid in "${PIDS[@]:-}"; do
        [ -n "$pid" ] || continue
        kill -TERM "$pid" 2>/dev/null || true
    done
    sleep 2
    for pid in "${PIDS[@]:-}"; do
        [ -n "$pid" ] || continue
        kill -KILL "$pid" 2>/dev/null || true
    done
}
trap cleanup EXIT INT TERM

client_log() { echo "$LOG_DIR/Client$1_$TODAY.log"; }

count_matches() {
    local pattern="$1" file="$2" n
    [ -f "$file" ] || { echo 0; return; }
    n=$(grep -cE -- "$pattern" "$file" 2>/dev/null)
    echo "${n:-0}"
}

last_session_count() {
    local n
    n=$(grep -o 'sessions=[0-9]*' "$MS_LOG" 2>/dev/null | tail -1 | cut -d= -f2)
    echo "${n:-0}"
}

wait_for() {
    local desc="$1" timeout="$2"; shift 2
    local waited=0
    while [ "$waited" -lt "$timeout" ]; do
        if "$@"; then return 0; fi
        sleep 1
        waited=$((waited + 1))
    done
    return 1
}

header "0. 사전 점검"

for binary in MainServer MainServerConsumer UserClient; do
    if [ -x "$BUILD_OUT/$binary" ]; then
        ok "$binary 존재"
    else
        fail "$BUILD_OUT/$binary 가 없다"
        info "먼저 빌드하라:"
        info "  CC=/opt/homebrew/opt/llvm/bin/clang CXX=/opt/homebrew/opt/llvm/bin/clang++ ./build.sh"
        exit 1
    fi
done

if ! command -v docker >/dev/null 2>&1; then
    fail "docker 가 필요하다 (Redis / RabbitMQ 기동용)"
    exit 1
fi
if ! command -v python3 >/dev/null 2>&1; then
    fail "python3 가 필요하다 (RabbitMQ 발행용)"
    exit 1
fi

for container in mainserver mainserver-consumer userclient; do
    if [ "$(docker inspect -f '{{.State.Running}}' "$container" 2>/dev/null)" = "true" ]; then
        docker stop "$container" >/dev/null 2>&1
        warn "도커 $container 를 중지했다 (로컬 바이너리와 포트/큐가 겹친다)"
    fi
done

docker compose -f "$COMPOSE" up -d redis rabbitmq >/dev/null 2>&1
if wait_for "redis" 30 docker exec redis redis-cli ping; then
    ok "Redis 준비 (localhost:6379)"
else
    fail "Redis 기동 실패"; exit 1
fi
if wait_for "rabbitmq" 60 docker exec rabbitmq rabbitmq-diagnostics -q ping; then
    ok "RabbitMQ 준비 (localhost:5672 / 15672)"
else
    fail "RabbitMQ 기동 실패"; exit 1
fi

if lsof -nP -iTCP:9876 -sTCP:LISTEN >/dev/null 2>&1; then
    fail "포트 9876 이 이미 점유되어 있다"
    lsof -nP -iTCP:9876 -sTCP:LISTEN 2>/dev/null | tail -n +2 | awk '{print "    " $1, $2}'
    exit 1
fi
ok "포트 9876 사용 가능"

header "1. 실행 환경 구성"

rm -rf "$RUN_DIR"
mkdir -p "$LOG_DIR"
for binary in MainServer MainServerConsumer UserClient; do
    cp "$BUILD_OUT/$binary" "$RUN_DIR/$binary"
done

python3 - "$RUN_DIR" <<'PYEOF'
import json, sys, pathlib

run_dir = pathlib.Path(sys.argv[1])

json.dump({
    "client_title": "MainServer",
    "high_priority_count": 3, "normal_priority_count": 3, "low_priority_count": 5,
    "write_interval": 100,
    "main_server_port": 9876, "buffer_size": 32768,
    "use_redis": True, "use_redis_tls": False,
    "redis_host": "127.0.0.1", "redis_port": 6379, "redis_ttl_sec": 600,
    "redis_db_global_message_index": 0,
}, open(run_dir / "main_server_configurations.json", "w"), indent=2)

json.dump({
    "client_title": "MainServerConsumer",
    "high_priority_count": 3, "normal_priority_count": 3, "low_priority_count": 5,
    "write_interval": 100,
    "use_redis": True, "use_redis_tls": False,
    "redis_host": "127.0.0.1", "redis_port": 6379, "redis_ttl_sec": 600,
    "redis_db_global_message_index": 0,
    "global_message_key": "send_global_message",
    "rabbit_mq_host": "localhost", "rabbit_mq_port": 5672,
    "rabbit_mq_user_name": "guest", "rabbit_mq_password": "guest",
    "consume_queue_name": "message_broadcast_queue",
    "use_database": False,
}, open(run_dir / "main_server_consumer_configurations.json", "w"), indent=2)

json.dump({
    "client_title": "UserClient",
    "high_priority_count": 3, "normal_priority_count": 3, "low_priority_count": 5,
    "write_interval": 100,
    "main_server_ip": "127.0.0.1", "main_server_port": 9876, "buffer_size": 32768,
}, open(run_dir / "user_client_configurations.json", "w"), indent=2)
PYEOF
ok "실행 디렉터리 구성: build/rtmc_test"
info "로그는 파일로 받는다 — 콘솔 stdout 은 블록 버퍼링되어 실행 중 확인이 안 된다"

header "2. 서버 기동"

"$RUN_DIR/MainServerConsumer" --write_file_log 4 --write_console_log 0 --log_root_path "$LOG_DIR/" &
PIDS+=("$!")
CONSUMER_PID="${PIDS[-1]}"

"$RUN_DIR/MainServer" --write_file_log 4 --write_console_log 0 --log_root_path "$LOG_DIR/" &
PIDS+=("$!")
MAINSERVER_PID="${PIDS[-1]}"

if wait_for "mainserver" 20 bash -c 'lsof -nP -iTCP:9876 -sTCP:LISTEN >/dev/null 2>&1'; then
    ok "MainServer 수신 대기 (9876)"
else
    fail "MainServer 가 9876 을 열지 못했다"
    tail -5 "$LOG_DIR/MainServer_$TODAY.log" 2>/dev/null | sed 's/^/    /'
    exit 1
fi

if wait_for "consumer" 30 bash -c "grep -q 'MainServerConsumer started successfully' '$LOG_DIR/MainServerConsumer_$TODAY.log' 2>/dev/null"; then
    ok "MainServerConsumer 큐 소비 시작"
else
    fail "MainServerConsumer 기동 실패"
    tail -5 "$LOG_DIR/MainServerConsumer_$TODAY.log" 2>/dev/null | sed 's/^/    /'
    exit 1
fi

header "3. 클라이언트 $CLIENTS 개 접속"

for i in $(seq 1 "$CLIENTS"); do
    "$RUN_DIR/UserClient" --client_title "Client$i" --write_file_log 4 --write_console_log 0 --log_root_path "$LOG_DIR/" &
    PIDS+=("$!")
done

MS_LOG="$LOG_DIR/MainServer_$TODAY.log"
for i in $(seq 1 "$CLIENTS"); do
    if wait_for "client$i" 20 bash -c "grep -q 'Received connection\[Client$i,' '$MS_LOG' 2>/dev/null"; then
        ok "Client$i 접속 (서버가 세션 수립을 기록)"
    else
        fail "Client$i 접속 실패"
        tail -4 "$(client_log "$i")" 2>/dev/null | sed 's/^/    /'
    fi
done

if wait_for "sessions" 20 bash -c "[ \"\$(grep -o 'sessions=[0-9]*' '$MS_LOG' 2>/dev/null | tail -1 | cut -d= -f2)\" = '$CLIENTS' ]"; then
    ok "서버 세션 수 = $(last_session_count) (maintenance 주기 기준)"
else
    fail "서버 세션 수 = $(last_session_count) (기대 $CLIENTS)"
fi

header "4. 브로드캐스트 $MESSAGES 건 발행"

TAG="rtmc_$(date +%s)"
PUBLISHED=0
for i in $(seq 1 "$MESSAGES"); do
    CODE=$(python3 - "$TAG" "$i" <<'PYEOF'
import json, sys, urllib.request, base64, uuid
tag, index = sys.argv[1], sys.argv[2]
payload = json.dumps({
    "id": "rtmc_test", "sub_id": f"s{index}",
    "publisher_information": {"message_id": str(uuid.uuid4()), "source": "rtmc_test"},
    "message": {"command": "chat_message", "content": f"{tag}_{index}"},
})
body = json.dumps({"properties": {}, "routing_key": "message_broadcast_queue",
                   "payload": payload, "payload_encoding": "string"}).encode()
req = urllib.request.Request("http://localhost:15672/api/exchanges/%2F/amq.default/publish",
                             data=body, method="POST",
                             headers={"content-type": "application/json",
                                      "authorization": "Basic " + base64.b64encode(b"guest:guest").decode()})
try:
    with urllib.request.urlopen(req, timeout=10) as resp:
        print(resp.status)
except Exception as err:
    print(f"ERR {err}")
PYEOF
)
    [ "$CODE" = "200" ] && PUBLISHED=$((PUBLISHED + 1))
done

if [ "$PUBLISHED" -eq "$MESSAGES" ]; then
    ok "RabbitMQ 발행 $PUBLISHED/$MESSAGES"
else
    fail "RabbitMQ 발행 $PUBLISHED/$MESSAGES"
fi

header "5. 클라이언트별 수신 확인"

count_received() { count_matches "$TAG" "$(client_log "$1")"; }

waited=0
while [ "$waited" -lt "$DELIVER_TIMEOUT" ]; do
    done_all=1
    for i in $(seq 1 "$CLIENTS"); do
        [ "$(count_received "$i")" -ge "$MESSAGES" ] || done_all=0
    done
    [ "$done_all" -eq 1 ] && break
    sleep 1
    waited=$((waited + 1))
done

TOTAL_RECEIVED=0
for i in $(seq 1 "$CLIENTS"); do
    RECEIVED=$(count_received "$i")
    TOTAL_RECEIVED=$((TOTAL_RECEIVED + RECEIVED))
    if [ "$RECEIVED" -ge "$MESSAGES" ]; then
        ok "Client$i 수신 $RECEIVED/$MESSAGES"
    else
        fail "Client$i 수신 $RECEIVED/$MESSAGES"
    fi
done
info "합계 $TOTAL_RECEIVED / $((MESSAGES * CLIENTS)) (메시지 $MESSAGES × 클라이언트 $CLIENTS)"

echo
echo "  수신 로그 표본 (Client1):"
grep "$TAG" "$(client_log 1)" 2>/dev/null | head -2 | sed 's/^/    /'

BROADCASTS=$(count_matches "send_broadcast_message" "$MS_LOG")
if [ "$BROADCASTS" -ge "$MESSAGES" ]; then
    ok "MainServer 브로드캐스트 발행 $BROADCASTS 건"
else
    fail "MainServer 브로드캐스트 발행 $BROADCASTS 건 (기대 $MESSAGES 이상)"
fi

LLEN=$(docker exec redis redis-cli LLEN send_global_message 2>/dev/null | tr -d '\r')
if [ "${LLEN:-1}" = "0" ]; then
    ok "Redis 큐 드레인 완료 (LLEN=0)"
else
    fail "Redis 큐 잔량 LLEN=${LLEN:-?}"
fi

header "6. 종료"

if [ "$KEEP_RUNNING" = "1" ] && [ "$FAILED" -eq 0 ]; then
    warn "KEEP_RUNNING=1 — 프로세스를 남겨 둔다"
    info "MainServer=$MAINSERVER_PID Consumer=$CONSUMER_PID"
    info "로그: $LOG_DIR"
    info "정리: kill ${PIDS[*]}"
else
    for pid in "${PIDS[@]}"; do
        kill -TERM "$pid" 2>/dev/null || true
    done
    ABNORMAL=0
    for pid in "${PIDS[@]}"; do
        rc=0
        wait "$pid" 2>/dev/null || rc=$?
        [ "$rc" -eq 0 ] || { ABNORMAL=$((ABNORMAL + 1)); info "PID $pid 종료 코드 $rc"; }
    done
    PIDS=()
    if [ "$ABNORMAL" -eq 0 ]; then
        ok "전 프로세스 정상 종료 (exit 0)"
    else
        fail "$ABNORMAL 개 프로세스가 비정상 종료"
    fi

    CRASH=0
    for f in "$LOG_DIR"/*.log; do
        [ -f "$f" ] || continue
        n=$(count_matches 'SIGSEGV|SIGABRT|terminate|future_error|bad_expected_access' "$f")
        CRASH=$((CRASH + n))
    done
    if [ "$CRASH" -eq 0 ]; then
        ok "크래시 마커 0건"
    else
        fail "크래시 마커 $CRASH 건"
    fi
fi

header "결과"
echo
echo -e "  통과: ${GREEN}$PASSED${NC}"
echo -e "  실패: ${RED}$FAILED${NC}"
echo
echo "  로그 디렉터리: $LOG_DIR"
echo
if [ "$FAILED" -eq 0 ]; then
    echo -e "${GREEN}✓ 브로드캐스트가 클라이언트 $CLIENTS 개 전체에 전달되었다${NC}"
    exit 0
fi
echo -e "${RED}✗ 실패 $FAILED 건${NC}"
exit 1
