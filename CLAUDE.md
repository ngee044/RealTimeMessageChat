# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

RealTimeMessageChat is a high-performance real-time messaging system built in C++17/20 using Boost.Asio for asynchronous I/O, RabbitMQ (AMQP) for message brokering, and Redis for caching. The system demonstrates a multi-process architecture where messages flow from an API server through RabbitMQ to a main server, which validates and broadcasts them to all connected TCP clients.

### Architecture Components

1. **MainServer** (`MainServer/`): TCP server that accepts client connections using `NetworkServer` from cpp_tool_kit. Handles client authentication, broadcasts messages received from RabbitMQ to all connected clients, and periodically syncs with Redis for state management.

2. **MainServerConsumer** (`MainServerConsumer/`): RabbitMQ consumer process that subscribes to message queues via `WorkQueueConsume`. It validates incoming messages, performs business logic, interacts with Redis, and publishes validated data for MainServer to broadcast.

3. **UserClient** (`UserClient/`): TCP client application using `NetworkClient` that connects to MainServer, receives broadcast messages, and sends user status updates.

4. **CommonModule** (`CommonModule/`): Shared message parsing and execution logic used by both server and client components. Contains:
   - `ServerMessageParsing/Execute`: Server-side message handlers
   - `ClientMessageParsing/Execute`: Client-side message handlers
   - `ServerCombinedMessageParsing/Execute`: Combined message handlers for server
   - `ClientCombinedMessageParsing/Execute`: Combined message handlers for client
   - `ModuleHeader.hpp`: Type aliases for callback functions

5. **cpp_tool_kit** (`.cpp_tool_kit/`): Submodule containing reusable C++ utilities:
   - `Network/`: `NetworkServer`, `NetworkClient`, `NetworkSession` - Boost.Asio based TCP networking
   - `Redis/`: `RedisClient`, `RedisConnector` - Redis integration with TLS support
   - `RabbitMQ/`: `WorkQueueConsume` - AMQP message queue consumer
   - `ThreadPool/`: Thread pool implementation for concurrent task execution
   - `Utilities/`: Logger, ArgumentParser, and other common utilities

## Build System

### Local Development (macOS/Linux)

```bash
# Build using vcpkg for dependency management
./build.sh

# Manual build (if build.sh doesn't work)
rm -rf build
mkdir build
cd build
cmake .. -DCMAKE_TOOLCHAIN_FILE="../../vcpkg/scripts/buildsystems/vcpkg.cmake" \
         -DCMAKE_BUILD_TYPE=Release \
         -DBUILD_SHARED_LIBS=OFF
cmake --build . --config Release --parallel
```

**Note**: The build script assumes vcpkg is located at `../../vcpkg` relative to the project root. Adjust the path in `build.sh` if your vcpkg installation is elsewhere.

Build outputs are located in:
- Executables: `build/out/`
- Libraries: `build/lib/`

### Docker Build & Run

```bash
cd docker
./docker-compose.sh
```

This builds and starts all services (RabbitMQ, Redis, MainServer, MainServerConsumer) using `docker/docker-compose.yml`.

**Docker management commands:**
```bash
# View logs
docker compose -f docker/docker-compose.yml logs -f mainserver
docker compose -f docker/docker-compose.yml logs -f mainserver-consumer

# Stop all services
docker compose -f docker/docker-compose.yml down

# Check service status
docker compose -f docker/docker-compose.yml ps
```

### Dependencies (vcpkg)

The project uses vcpkg for dependency management. Key dependencies defined in `vcpkg.json`:
- `boost-asio`, `boost-json`, `boost-system`, `boost-filesystem`, `boost-process`
- `librabbitmq` - RabbitMQ AMQP client
- `redis-plus-plus` - Redis client with C++17 and TLS support
- `libpq` - PostgreSQL client library
- `cryptopp` - Cryptography library
- `fmt` - String formatting
- `gtest` - Google Test framework
- `lz4` - Compression

## Configuration Files

Each component has a JSON configuration file:
- `MainServer/main_server_configurations.json`: Server IP/port, Redis connection, RabbitMQ queue settings, thread pool sizes, SSL/TLS options
- `MainServerConsumer/main_server_consumer_configurations.json`: RabbitMQ consumer settings, Redis connection, thread pool configuration
- `UserClient/user_client_configurations.json`: Server connection details, client behavior settings

For Docker deployments, configuration files are located in `docker/config/` and mounted into containers.

**Key configuration parameters:**
- `server_ip`, `server_port`: MainServer TCP listening address
- `redis_host`, `redis_port`, `redis_db_*_index`: Redis connection and database indices
- `use_redis`, `use_redis_tls`: Enable Redis and TLS
- `consume_queue_name`: RabbitMQ queue name for MainServerConsumer
- `high_priority_count`, `normal_priority_count`, `low_priority_count`: Thread pool worker counts
- `buffer_size`: Network buffer size for TCP connections
- `encrypt_mode`, `use_ssl`: Enable encryption and SSL/TLS

## Key Design Patterns

### Message Flow
1. API server publishes message to RabbitMQ
2. MainServerConsumer subscribes and consumes message
3. Consumer validates, performs business logic with Redis
4. Consumer stores result in Redis with a known key
5. MainServer polls Redis for new messages
6. MainServer broadcasts to all connected TCP clients via NetworkServer

### Callback-Based Message Handling
Both server and client components use callback maps defined in `ModuleHeader.hpp`:
- `server_message_execute_callback`: `(const std::string& message) -> std::tuple<bool, std::optional<std::string>>`
- `client_message_parsing_callback`: `(const std::string& id, sub_id, command, message) -> std::tuple<bool, std::optional<std::string>>`

Message handlers are registered in a `std::map` where the key is the message command string, and the value is the callback function.

### Redis Usage
- **User Status Cache**: `redis_db_user_status_index` - stores online/offline status
- **Global Messages**: `redis_db_global_message_index` - used for broadcasting
- **TTL**: Configurable via `redis_ttl_sec` for automatic cache expiration

### Thread Pool
Each component creates a thread pool with three priority queues:
- High priority: Critical tasks (connection handling)
- Normal priority: Standard message processing
- Low priority: Background jobs (periodic DB sync)

## Running Tests

The project includes Google Test (gtest) for unit testing. Tests are typically located in subdirectories or the `.cpp_tool_kit` submodule.

```bash
# Build with tests enabled
cd build
cmake --build . --target all

# Run tests (if test executables are built)
ctest --output-on-failure
```

## Common Development Scenarios

### Adding a New Message Type
1. Define the message structure in JSON format
2. Add callback handler in appropriate `*MessageParsing.cpp` and `*MessageExecute.cpp` in `CommonModule/`
3. Register the callback in the server or client's message map with a unique command key
4. Update the `Configurations` class if new configuration parameters are needed

### Working with Redis
- Use `RedisClient` from `.cpp_tool_kit/Redis/`
- Each component can have multiple Redis client instances for different DB indices
- Always check `use_redis()` configuration before Redis operations
- Use `redis_ttl_sec()` for setting expiration times

### Modifying Network Behavior
- Server-side: Modify `MainServer::received_connection()` and `MainServer::received_message()`
- Client-side: Modify `UserClient::received_connection()` and `UserClient::received_message()`
- Network layer is abstracted in `NetworkServer` and `NetworkClient` from cpp_tool_kit

### Adding RabbitMQ Consumers
- Extend `MainServerConsumer` or create a new consumer component
- Use `WorkQueueConsume` from `.cpp_tool_kit/RabbitMQ/`
- Define queue name in configuration JSON
- Implement consume callback to handle incoming AMQP messages

## Debugging Tips

### Logging
All components use `Logger` from `.cpp_tool_kit/Utilities/`:
```cpp
Logger::handle().write(LogTypes::Information, "Message");
Logger::handle().write(LogTypes::Error, "Error message");
```

Log files are written to the path specified by `log_root_path()` in configurations.

Enable console and file logging via:
- `write_console()`: Set log level for console output
- `write_file()`: Set log level for file output
- `write_interval()`: Milliseconds between log flushes

### Docker Debugging
```bash
# Enter running container
docker exec -it mainserver /bin/bash

# Check if services are reachable
# Inside container:
nc -zv rabbitmq 5672
nc -zv redis 6379
```

## Project Conventions

- **Return Types**: Most functions return `std::tuple<bool, std::optional<std::string>>` where the bool indicates success, and the optional string contains an error message on failure.
- **C++ Standard**: C++17 minimum, C++20 enabled (CMAKE_CXX_STANDARD 20)
- **Naming**: Snake_case for variables/functions, PascalCase for classes
- **Error Handling**: No exceptions in network/async code paths; use return tuples
- **Thread Safety**: Use `std::mutex` for shared state (see `MainServer::mutex_`, `UserClient::mutex_`)
