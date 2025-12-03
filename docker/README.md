# RealTimeMessageChat Docker 환경

Docker를 통해 MainServer, MainServerConsumer, RabbitMQ, Redis를 실행하는 가이드입니다.

## 아키텍처

```
┌─────────────┐
│ UserClient  │ (로컬에서 실행)
└──────┬──────┘
       │ TCP/IP (port 9876)
       ▼
┌─────────────────────────────────────────────────┐
│              Docker Environment                  │
│                                                  │
│  ┌─────────────┐         ┌──────────────┐      │
│  │ MainServer  │────────▶│   RabbitMQ   │      │
│  │  (port 9876)│         │  (port 5672) │      │
│  └─────────────┘         └──────┬───────┘      │
│         │                        │              │
│         │                        │              │
│         ▼                        ▼              │
│  ┌─────────────┐         ┌──────────────┐      │
│  │    Redis    │◀────────│MainServer    │      │
│  │  (port 6379)│         │  Consumer    │      │
│  └─────────────┘         └──────────────┘      │
│                                                  │
└─────────────────────────────────────────────────┘
```

## 빠른 시작

### 1. 빌드

```bash
./docker/build.sh
```

### 2. 시작

```bash
./docker/start.sh
```

### 3. 로그 확인

```bash
# 모든 서비스 로그
./docker/logs.sh

# 특정 서비스 로그
./docker/logs.sh mainserver
./docker/logs.sh mainserver-consumer
./docker/logs.sh rabbitmq
./docker/logs.sh redis
```

### 4. 중지

```bash
./docker/stop.sh
```

### 5. 완전 정리 (이미지, 볼륨 포함)

```bash
./docker/clean.sh
```

## UserClient 연결 방법

UserClient는 로컬에서 실행하고 Docker의 MainServer에 연결합니다:

### 방법 1: UserClient 빌드 및 실행

```bash
# 프로젝트 루트에서
mkdir -p build && cd build
cmake ..
make UserClient

# UserClient 실행
./out/UserClient
```

### 방법 2: 기존 빌드된 UserClient 사용

UserClient의 설정 파일에서 MainServer 주소가 `localhost:9876` 또는 `127.0.0.1:9876`으로 설정되어 있는지 확인하세요.

## 서비스 정보

### 포트

- **MainServer (TCP)**: `localhost:9876`
- **Redis**: `localhost:6379`
- **RabbitMQ (AMQP)**: `localhost:5672`
- **RabbitMQ Management UI**: `http://localhost:15672`
  - Username: `guest`
  - Password: `guest`

### 서비스 상태 확인

```bash
docker compose -f docker/docker-compose.yml ps
```

### 개별 서비스 재시작

```bash
# MainServer 재시작
docker compose -f docker/docker-compose.yml restart mainserver

# Consumer 재시작
docker compose -f docker/docker-compose.yml restart mainserver-consumer
```

## 테스트 시나리오

1. **Docker 환경 시작**
   ```bash
   ./docker/start.sh
   ```

2. **서비스 상태 확인**
   ```bash
   docker compose -f docker/docker-compose.yml ps
   # 모든 서비스가 "Up" 상태인지 확인
   ```

3. **로그 모니터링 시작** (새 터미널)
   ```bash
   ./docker/logs.sh
   ```

4. **UserClient 실행** (새 터미널)
   ```bash
   cd build
   ./out/UserClient
   ```

5. **메시지 전송 테스트**
   - UserClient에서 메시지 입력
   - MainServer가 메시지를 받아 RabbitMQ에 발행
   - Consumer가 RabbitMQ에서 메시지를 소비
   - 로그에서 메시지 흐름 확인

6. **RabbitMQ Management UI 확인**
   - 브라우저에서 `http://localhost:15672` 접속
   - `guest` / `guest`로 로그인
   - Queues 탭에서 `message_broadcast_queue` 확인

## 트러블슈팅

### 포트가 이미 사용 중인 경우

```bash
# 포트 사용 확인
lsof -i :9876
lsof -i :6379
lsof -i :5672
lsof -i :15672

# 필요시 프로세스 종료
kill -9 <PID>
```

### 컨테이너가 시작되지 않는 경우

```bash
# 로그 확인
docker compose -f docker/docker-compose.yml logs mainserver
docker compose -f docker/docker-compose.yml logs mainserver-consumer

# 컨테이너 재빌드
./docker/clean.sh
./docker/build.sh
./docker/start.sh
```

### 설정 파일 수정

설정을 변경하려면 `docker/docker-compose.yml`의 environment 섹션을 수정:

```yaml
environment:
  REDIS_HOST: redis
  REDIS_PORT: 6379
  RABBITMQ_HOST: rabbitmq
  RABBITMQ_PORT: 5672
  # ... 기타 설정
```

## 주의사항

- Docker Desktop이 실행 중이어야 합니다
- 최소 요구사항: Docker 20.10+, Docker Compose 1.29+
- 첫 빌드는 vcpkg 의존성 설치로 인해 시간이 오래 걸릴 수 있습니다 (10-30분)
- UserClient는 Docker 외부(로컬)에서 실행되어야 합니다
