# RealTimeMessageChat
**RealTimeMessageChat**은 다음과 같은 목표와 기술 스택을 통해 **실시간 메시징**을 구현한 프로젝트입니다.

1. **고성능 서버**:
    - **C++17**과 **Boost.Asio**의 `io_context`를 사용해 비동기 I/O를 처리하며, TCP/IP 기반으로 **실시간** 통신을 지원합니다.
2. **메시지 브로커**:
    - *RabbitMQ(AMQP)**를 통해 **API 서버**와 **메인 서버** 간 메시지 발행(Publish)·구독(Consume)을 수행합니다.
3. **데이터베이스 & 캐싱**:
    - **Redis**를 적용하여 서버와 DB 간 지연을 최소화하고, 비즈니스 로직 수행 시 필요한 데이터를 빠르게 검증·처리합니다.
4. **브로드캐스팅**:
    - 메인 서버는 RabbitMQ에서 메시지를 소비한 뒤 검증 및 비즈니스 로직을 수행하고, **모든 클라이언트(메시지를 발행한 클라이언트 포함)**에게 **TCP** 기반으로 메시지를 브로드캐스트합니다.
5. **확장성**:
    - 다중 프로세스·멀티 스레드 아키텍처를 기반으로, 향후 다른 통신 규약이나 도메인 요구사항에도 유연하게 확장할 수 있도록 설계했습니다.
    

이 프로젝트는 **REST API** 요청에 의해 시작된 메시지가 AMQP 메시지 큐를 거쳐, **메인 서버**에서 **검증 및 처리 후 연결된 모든 클라이언트**에게 실시간으로 전달되는 **프로토타입**이자 **기술 시연용**입니다.

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
