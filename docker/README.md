# Docker 기반 RealTimeMessageChat 테스트 환경

이 디렉토리는 RealTimeMessageChat 시스템을 Docker로 배포하고 로컬 UserClient로 테스트하기 위한 스크립트와 설정 파일을 포함합니다.

## 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Environment                        │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────┐  │
│  │   RabbitMQ   │  │    Redis     │  │   MainServer      │  │
│  │  :5672/15672 │  │    :6379     │  │     :9876         │  │
│  └──────┬───────┘  └──────┬───────┘  └─────────┬─────────┘  │
│         │                 │                     │            │
│         │      ┌──────────┴─────────┐          │            │
│         │      │                    │          │            │
│         └──────┤ MainServerConsumer ├──────────┘            │
│                │                    │                        │
│                └────────────────────┘                        │
└─────────────────────────────────────┬───────────────────────┘
                                      │ TCP :9876
                                      │
┌─────────────────────────────────────┴───────────────────────┐
│                    Local Machine                             │
│                                                              │
│              ┌────────────────────┐                          │
│              │     UserClient     │                          │
│              │   (localhost:9876) │                          │
│              └────────────────────┘                          │
└──────────────────────────────────────────────────────────────┘
```

## 사전 요구사항

- Docker 및 Docker Compose 설치
- 프로젝트 빌드 완료 (UserClient 실행을 위해)

## 빠른 시작

### 1. Docker 서비스 시작

```bash
cd docker
./docker-compose.sh
```

이 명령은 다음 서비스를 시작합니다:
- **RabbitMQ**: 메시지 브로커 (관리 UI: http://localhost:15672, guest/guest)
- **Redis**: 캐시 서버
- **MainServer**: TCP 서버 (포트 9876)
- **MainServerConsumer**: RabbitMQ 메시지 소비자

### 2. 서비스 상태 확인

```bash
./status.sh
```

### 3. UserClient 실행

**먼저 프로젝트를 빌드해야 합니다:**
```bash
cd ..
mkdir -p build && cd build
cmake .. -DCMAKE_TOOLCHAIN_FILE="../../vcpkg/scripts/buildsystems/vcpkg.cmake" -DCMAKE_BUILD_TYPE=Release
cmake --build . --config Release
cd ../docker
```

**UserClient 실행:**
```bash
./run-client.sh
```

### 4. 테스트 메시지 발행

RabbitMQ에 테스트 메시지를 발행하여 브로드캐스트 테스트:

```bash
./publish-message.sh "Hello World!"
```

### 5. 로그 확인

```bash
# 모든 서비스 로그
./logs.sh all

# 특정 서비스 로그
./logs.sh mainserver
./logs.sh consumer
./logs.sh rabbitmq
./logs.sh redis
```

### 6. 서비스 중지

```bash
./stop.sh
```

## 스크립트 설명

| 스크립트 | 설명 |
|---------|------|
| `docker-compose.sh` | Docker 서비스 빌드 및 시작 |
| `stop.sh` | 모든 Docker 서비스 중지 |
| `status.sh` | 서비스 상태 및 헬스 체크 |
| `logs.sh` | 서비스 로그 확인 |
| `run-client.sh` | 로컬 UserClient 실행 |
| `publish-message.sh` | RabbitMQ에 테스트 메시지 발행 |

## 스크립트 사용법 요약

- `docker-compose.sh`: RabbitMQ, Redis, MainServer, MainServerConsumer를 빌드 후 실행합니다. `docker-compose.yml`의 `args.VCPKG_TRIPLET`를 필요 시 수정하세요(예: `arm64-linux`).
- `status.sh`: 컨테이너 상태와 헬스체크를 표시합니다. 실행 중인지 빠르게 확인할 때 사용합니다.
- `logs.sh [target]`: `all`, `mainserver`, `consumer`, `rabbitmq`, `redis` 중 하나를 지정해 해당 로그를 스트리밍합니다.
- `stop.sh`: 모든 컨테이너를 중지합니다. 테스트 종료 시 사용합니다.
- `run-client.sh`: 로컬 빌드된 `UserClient`를 실행합니다. 실행 전 상위 디렉터리에서 프로젝트를 빌드해야 합니다.
- `publish-message.sh "메시지 내용"`: RabbitMQ에 테스트 메시지를 발행하여 브로드캐스트 경로를 검증합니다.

## 설정 파일

| 파일 | 설명 |
|------|------|
| `config/main_server_configurations.json` | MainServer 설정 |
| `config/main_server_consumer_configurations.json` | MainServerConsumer 설정 |
| `config/user_client_configurations.json` | UserClient 설정 (로컬 실행용) |

## 포트 매핑

| 서비스 | 포트 | 설명 |
|--------|------|------|
| RabbitMQ AMQP | 5672 | 메시지 큐 프로토콜 |
| RabbitMQ Management | 15672 | 웹 관리 UI |
| Redis | 6379 | 캐시 서버 |
| MainServer | 9876 | TCP 클라이언트 연결 |

## 문제 해결

### MainServer에 연결할 수 없음

1. Docker 서비스가 실행 중인지 확인:
   ```bash
   ./status.sh
   ```

2. MainServer 로그 확인:
   ```bash
   ./logs.sh mainserver
   ```

### 빌드 실패

ARM64 Mac에서 빌드 시 triplet 지정이 필요할 수 있습니다:
```bash
# docker-compose.yml에서 args 주석 해제
args:
  VCPKG_TRIPLET: arm64-linux
```

### RabbitMQ 연결 문제

RabbitMQ가 완전히 시작될 때까지 기다려야 합니다. `status.sh`로 헬스 체크를 확인하세요.
