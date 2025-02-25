# RealTimeMessageChat
 TCP/IP 기반의 실시간 대규모 메시지 처리 시스템
 포트폴리오 목적의 서버 설계 및 구현의 기본 내용 구현

## 기본 구성
```mermaid
sequenceDiagram
    participant Client as Client
    participant APIServer as API Server
    participant MQ as RabbitMQ
    participant MainServer as Main Server
    participant Redis as Redis
    participant All Clients as All Clients

    Client->>APIServer: 1) POST /publishMessage (JSON)
    note right of Client: 클라이언트에서 REST API 호출<br>메시지 JSON 전달

    APIServer->>MQ: 2) Publish 메시지
    note right of APIServer: API 서버가<br>RabbitMQ에 메시지 발행

    MainServer->>MQ: 3) Subscribe/Consume
    note right of MainServer: 메인 서버는<br>RabbitMQ 메시지 구독 중

    MQ-->>MainServer: 4) Deliver 메시지
    note right of MQ: 메시지를 받은<br>메인 서버

    MainServer->>Redis: 5) Validation/비즈니스 로직 수행<br>필요 시 Redis 연동
    note right of Redis: 서버와 DB 사이<br>지연 최소화를 위해 Redis 사용

    MainServer->>All Clients: 6) Broadcast 메시지 (TCP)
    note right of All Clients: Boost 기반<br>TCP 연결로 실시간 메시지 수신
```
---
### 클라이언트 접속 및 Redis 데이터 조회
```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Redis
    participant DB

    Client->>Server: TCP/IP 연결 요청
    Server-->>Client: 연결 수락
    Client->>Server: 데이터 요청 (예: 유저 상태)
    Server->>Redis: 캐시 조회 요청
    alt 캐시에 데이터가 존재함
        Redis-->>Server: 캐시 데이터 응답
        Server-->>Client: 데이터 응답 (빠른 응답)
    else 캐시에 데이터 없음
        Server->>DB: 데이터 조회 요청
        DB-->>Server: DB 데이터 응답
        Server->>Redis: 캐시 업데이트
        Server-->>Client: DB 데이터 응답
    end
```
---
### 메시지 큐를 이용한 브로드 캐스트 흐름
```mermaid
sequenceDiagram
    participant Client1
    participant Server
    participant MessageQueue
    participant Client2
    participant Client3

    Client1->>Server: 메시지 전송 (JSON)
    Server->>MessageQueue: 메시지 저장 (Publish)
    MessageQueue-->>Server: 메시지 브로드캐스트 (Subscribe)
    Server-->>Client2: 메시지 전달
    Server-->>Client3: 메시지 전달
```
---
### Redis → PostgreSQL 데이터 동기화 흐름
```mermaid
sequenceDiagram
    participant Scheduler
    participant Redis
    participant DB

    Scheduler->>Redis: 캐시 데이터 가져오기
    Redis-->>Scheduler: 캐시 데이터 응답
    Scheduler->>DB: DB 저장 요청 (Batch Insert/Update)
    DB-->>Scheduler: 저장 완료 응답
    Scheduler->>Redis: 해당 캐시 데이터 삭제 (Optional)
```
---
### 클라이언트↔서버 인증 및 보안 흐름
```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant DB
    participant Redis

    Client->>Server: 로그인 요청 (ID/PW)
    Server->>DB: 사용자 인증 정보 조회
    alt 유효한 사용자
        DB-->>Server: 인증 성공
        Server->>Redis: 세션 정보 저장 (JWT/토큰)
        Server-->>Client: 인증 성공 (토큰 발급)
    else 인증 실패
        DB-->>Server: 인증 실패 응답
        Server-->>Client: 로그인 실패 응답
    end
```
---
### 비정상 세션 종료 처리 흐름
```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Redis
    participant DB

    Client->>Server: 연결 유지 요청 (Heartbeat)
    alt 정상적인 연결 유지
        Server-->>Client: 응답 (OK)
    else 클라이언트 비정상 종료 (예: 네트워크 단절)
        Server->>Redis: 세션 정보 확인 (예: UserID)
        alt 세션 존재
            Redis-->>Server: 세션 존재 응답
            Server->>Redis: 세션 삭제 (DEL UserID)
            Server->>DB: 유저 상태 "오프라인"으로 업데이트
            DB-->>Server: 업데이트 완료 응답
        else 세션 없음 (이미 만료됨)
            Server-->>Server: 추가 처리 필요 없음
        end
    end
```
