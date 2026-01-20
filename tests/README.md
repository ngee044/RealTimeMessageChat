# RealTimeMessageChat Tests

This directory contains test scripts for validating the RealTimeMessageChat system.

## Test Scripts

### test-local.sh

Local test script that validates the built executables without requiring Docker services.

**What it tests:**
- Binary existence (MainServer, UserClient, MainServerConsumer)
- Configuration file detection
- MainServer startup and port listening (9876)
- UserClient connection to MainServer
- Graceful shutdown via SIGTERM signal handling

**Usage:**
```bash
cd tests
./test-local.sh
```

**Prerequisites:**
- Built executables in `build/out/`
- No Docker services required

### test-integration.sh

Full integration test script that validates the entire message flow with Docker services.

**What it tests:**
- Docker service health (RabbitMQ, Redis, PostgreSQL, MainServer)
- Redis SET/GET operations
- RabbitMQ queue declaration and message publishing
- UserClient connection to MainServer (via Docker)
- REST API health check and message sending
- Log accessibility for MainServer and Consumer

**Usage:**
```bash
# First, start Docker services
cd docker
./docker-compose.sh

# Then run integration tests
cd ../tests
./test-integration.sh
```

**Prerequisites:**
- Docker services running (`docker/docker-compose.sh`)
- Built executables in `build/out/` (for UserClient test)
- `jq` command available for JSON processing
- `nc` (netcat) command available for port checking

## Environment Variables

Both scripts support the following environment variables for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `RABBITMQ_HOST` | localhost | RabbitMQ server hostname |
| `RABBITMQ_PORT` | 15672 | RabbitMQ Management API port |
| `RABBITMQ_USER` | guest | RabbitMQ username |
| `RABBITMQ_PASS` | guest | RabbitMQ password |
| `QUEUE_NAME` | message_broadcast_queue | RabbitMQ queue name |

## Helper Scripts (in docker/)

### publish-message.sh

Publish a test message to RabbitMQ in the format expected by MainServerConsumer.

**Usage:**
```bash
cd docker

# Basic usage (default message)
./publish-message.sh

# Custom message
./publish-message.sh "Hello World!"

# Full customization: message, user_id, command
./publish-message.sh "My message" "user123" "chat_message"
```

### status.sh

Check the health status of all Docker services.

```bash
cd docker
./status.sh
```

### logs.sh

View logs from Docker services.

```bash
cd docker
./logs.sh all           # All services
./logs.sh mainserver    # MainServer only
./logs.sh consumer      # MainServerConsumer only
./logs.sh rabbitmq      # RabbitMQ only
./logs.sh redis         # Redis only
```

## Test Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                      Integration Test Flow                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Docker Services Check                                        │
│     └── RabbitMQ, Redis, PostgreSQL, MainServer                 │
│                                                                  │
│  2. Redis Operations Test                                        │
│     └── SET → GET → DEL                                         │
│                                                                  │
│  3. RabbitMQ Test                                                │
│     └── Queue Declaration → Message Publish                     │
│                                                                  │
│  4. UserClient Connection Test                                   │
│     └── Connect to MainServer → Verify Connection               │
│                                                                  │
│  5. REST API Test                                                │
│     └── Health Check → Send Message API                         │
│                                                                  │
│  6. Log Verification                                             │
│     └── MainServer logs → Consumer logs                         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Troubleshooting

### Common Issues

1. **"Docker services are not running"**
   - Start Docker services: `cd docker && ./docker-compose.sh`

2. **"Binary not found"**
   - Build the project: `cd .. && ./build.sh`

3. **"nc: command not found"**
   - Install netcat: `brew install netcat` (macOS) or `apt install netcat` (Linux)

4. **"jq: command not found"**
   - Install jq: `brew install jq` (macOS) or `apt install jq` (Linux)

5. **"Connection refused on port 9876"**
   - Check if MainServer is running: `docker ps | grep mainserver`
   - Check MainServer logs: `cd docker && ./logs.sh mainserver`
