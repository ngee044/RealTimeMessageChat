# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

RealTimeMessageChat is a real-time messaging system using C++23, Boost.Asio, RabbitMQ, Redis, and PostgreSQL. Multi-process architecture: REST API → RabbitMQ → Consumer → Redis/PostgreSQL → TCP Server → Clients.

### Architecture Components

1. **RestAPI** (`RestAPI/`): Go (Gin) REST API. Entry point for messages — publishes to RabbitMQ. Has JWT auth, Prometheus metrics, Swagger docs.
2. **MainServer** (`MainServer/`): Boost.Asio TCP server. Polls Redis for new messages, broadcasts to connected clients.
3. **MainServerConsumer** (`MainServerConsumer/`): RabbitMQ consumer. Validates messages, stores to Redis + PostgreSQL (via `DBWorker`).
4. **UserClient** (`UserClient/`): TCP client connecting to MainServer.
5. **CommonModule** (`CommonModule/`): Shared message parsing/execution callbacks. 8 callback type aliases in `ModuleHeader.hpp` — all return `std::tuple<bool, std::optional<std::string>>`.
6. **cpp_tool_kit** (`.cpp_tool_kit/`): Git submodule — `NetworkServer`/`NetworkClient`, `RedisClient`, `WorkQueueConsume`, `ThreadPool`, `Logger`, etc.
7. **database** (`database/`): PostgreSQL schema (`schema.sql`) — `messages` table, `recent_messages` view, `cleanup_old_messages()`. Optional AES-256-CBC encryption.

## Build & Run

### C++ (Local)

```bash
# build.sh only runs cmake configure — you must also run the build step:
./build.sh
cmake --build build --config Release --parallel

# Or full manual build (recommended — build.sh has a typo in -DCMAKE_BUILD_TYPE):
rm -rf build && mkdir build && cd build
cmake .. -DCMAKE_TOOLCHAIN_FILE="$HOME/vcpkg/scripts/buildsystems/vcpkg.cmake" \
         -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF
cmake --build . --config Release --parallel
```

Outputs: `build/out/` (MainServer, UserClient, MainServerConsumer), `build/lib/`

### RestAPI (Go)

```bash
cd RestAPI
make run                          # Run with default config
make build                        # Build binary
make test                         # Tests with race detection + coverage
make check                        # fmt + vet + lint + test (full validation)
make swagger                      # Regenerate Swagger docs
make install-tools                # Install golangci-lint + swag
make dev                          # Auto-reload with air
```

### Docker

```bash
cd docker
./docker-compose.sh               # Build and start all services
./status.sh                       # Health check
./logs.sh all                     # Logs (also: mainserver, consumer, rabbitmq, redis)
./stop.sh                         # Stop all
./run-client.sh                   # Run local UserClient → Docker MainServer
./publish-message.sh "Hello!"     # Test message via RabbitMQ
```

**Service ports:** MainServer 9876 (TCP), RestAPI 8080 (HTTP, health at `/health`), RabbitMQ 5672/15672 (AMQP/Management), Redis 6379, PostgreSQL 5432 (DB: rtmc, User: rtmc_user).

## Running Tests

```bash
# C++ unit tests (gtest)
cd build && ctest --output-on-failure

# Go unit tests (single package: go test -v -race ./internal/handlers/...)
cd RestAPI && make test

# Local smoke tests (binary startup, port binding, SIGTERM handling)
./tests/test-local.sh

# Integration tests (requires Docker services running via docker-compose.sh)
./tests/test-integration.sh
```

## CI/CD

No CI is configured — there are no GitHub Actions workflows. Verification is local only:
`./scripts/syntax-check.sh`, `docker compose build`, `./tests/test-local.sh`, `./tests/test-integration.sh`.

## Configuration

Each component has a JSON config file — refer to these directly for available parameters:
- `MainServer/main_server_configurations.json` — server IP/port, Redis, thread pool, SSL
- `MainServerConsumer/main_server_consumer_configurations.json` — RabbitMQ, Redis, PostgreSQL, DB encryption
- `UserClient/user_client_configurations.json` — server connection
- `RestAPI/config/api_server_config.json` — server, RabbitMQ, Redis, PostgreSQL, JWT, metrics, logging

Docker configs: `docker/config/` (mounted read-only into containers).

## Key Design Patterns

### Message Flow
```
Client HTTP POST → RestAPI → RabbitMQ → MainServerConsumer
                                         ↓ validate + business logic
                                      Redis (cache) + PostgreSQL (persist via DBWorker)
                                         ↓
                                      MainServer (polls Redis) → TCP broadcast to clients
```

### RestAPI Layered Architecture
`cmd/server/main.go` → `internal/handlers/` → `internal/service/` → `internal/repository/`

Handlers: `message_handler.go` (basic RabbitMQ publish), `message_handler_extended.go` (DB-backed CRUD), `user_handler.go`, `system_handler.go`. Middleware: auth (JWT), logger, metrics (Prometheus), rate limiting (10 req/s, burst 20). Extended message and user endpoints are conditionally registered only when PostgreSQL is available. Infrastructure services (RabbitMQ, Redis, PostgreSQL) are initialized in `internal/services/`.

### Callback-Based Message Handling
Message handlers are registered in `std::map<string, callback>` where each command string maps to a handler function. `CommonModule/ModuleHeader.hpp` defines 8 callback types in two families:
- **Server-side**: `server_message_execute_callback`, `server_message_parsing_callback`, `server_combine_message_callback`, `server_combine_message_parsing_callback`
- **Client-side**: `client_message_execute_callback`, `client_message_parsing_callback`, `client_combine_message_execute_callback`, `client_combine_message_parsing_callback`

"Combine" variants accept a `std::vector<uint8_t>` for binary data. All return `std::tuple<bool, std::optional<std::string>>`.

### Adding a New Message Type
1. Add callback handler in `CommonModule/*MessageParsing.cpp` and `*MessageExecute.cpp`
2. Register in the component's callback map with a unique command key
3. Update `Configurations` class if new config parameters are needed

### Thread Pool
Three priority queues per component: high (connection handling), normal (message processing), low (background DB sync). Configured via `high_priority_count`, `normal_priority_count`, `low_priority_count` in JSON config.

## Project Conventions

- **C++ Standard**: C++23 (`CMAKE_CXX_STANDARD 23`), CMake 3.18+
- **Go**: 1.23+ for RestAPI
- **Formatting**: `.clang-format` in `.cpp_tool_kit/` — GNU-based, tabs, 4-space width, 170 column limit
- **Naming**: PascalCase for classes/files, snake_case for functions/variables, `member_` suffix for class members
- **Return Types**: `std::tuple<bool, std::optional<std::string>>` — bool for success, optional string for error message
- **Error Handling**: No exceptions in network/async code paths; use return tuples
- **Thread Safety**: `std::mutex` for shared state (`MainServer::mutex_`, `UserClient::mutex_`)
- **Dependencies**: vcpkg (`~/vcpkg/scripts/buildsystems/vcpkg.cmake`). Key: boost-asio, boost-json, librabbitmq, redis-plus-plus, libpq, cryptopp, gtest, lz4
- **Submodule**: `git submodule update --init --recursive` for `.cpp_tool_kit/`
