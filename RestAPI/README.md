# RealTimeMessageChat REST API

A production-grade REST API service built with Go and Gin framework that publishes messages to RabbitMQ for the RealTimeMessageChat system.

## Features

- **High Performance**: Built with Go and Gin framework for optimal performance
- **Message Publishing**: Publishes messages to RabbitMQ with priority in payload
- **Auto-Reconnection**: Automatic reconnection to RabbitMQ on connection loss
- **Structured Logging**: JSON-formatted logging with multiple levels
- **Health Checks**: Built-in health check endpoints for monitoring
- **Graceful Shutdown**: Clean shutdown with connection draining
- **Docker Support**: Containerized deployment with health checks
- **CORS Support**: Cross-Origin Resource Sharing enabled
- **Input Validation**: Comprehensive request validation
- **Error Handling**: Robust error handling with detailed error responses

## Architecture

The REST API serves as the entry point for the RealTimeMessageChat system:

```
Client Request → REST API → RabbitMQ → MainServerConsumer → Redis → MainServer → TCP Clients
```

## Project Structure

```
RestAPI/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handlers/
│   │   └── message_handler.go  # HTTP request handlers
│   ├── middleware/
│   │   └── logger.go            # HTTP middleware (logging, CORS, recovery)
│   ├── models/
│   │   └── message.go           # Data models and validation
│   └── services/
│       └── rabbitmq.go          # RabbitMQ service layer
├── pkg/
│   └── logger/
│       └── logger.go            # Logging utilities
├── config/
│   └── api_server_config.json  # Configuration file
├── Dockerfile                   # Container image definition
├── go.mod                       # Go module dependencies
└── README.md                    # This file
```

## Prerequisites

- Go 1.23 or later
- RabbitMQ server
- Docker (optional, for containerized deployment)

## Installation

### Local Development

1. **Clone the repository**
   ```bash
   cd RealTimeMessageChat/RestAPI
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure the application**

   Edit `config/api_server_config.json`:
   ```json
   {
     "server": {
       "host": "0.0.0.0",
       "port": 8080,
       "mode": "release"
     },
     "rabbitmq": {
       "host": "localhost",
       "port": 5672,
       "username": "guest",
       "password": "guest",
       "queue_name": "message_queue"
     },
     "logging": {
       "level": "info",
       "format": "json",
       "output_path": "logs/api_server.log"
     }
   }
   ```

4. **Build the application**
   ```bash
   go build -o api-server ./cmd/server
   ```

5. **Run the application**
   ```bash
   ./api-server -config config/api_server_config.json
   ```

### Docker Deployment

1. **Build the Docker image**
   ```bash
   docker build -t rtmc-api-server:latest .
   ```

2. **Run with docker-compose** (from project root)
   ```bash
   cd docker
   docker-compose up api-server
   ```

## API Endpoints

### POST /api/v1/send_message

Send a message to RabbitMQ for processing.

**Request Body:**
```json
{
  "user_id": "user123",
  "command": "chat_message",
  "sub_id": "room001",
  "content": "Hello, World!",
  "metadata": {
    "room_name": "General",
    "timestamp": 1234567890
  },
  "priority": 2
}
```

**Fields:**
- `user_id` (required): Unique identifier for the user
- `command` (required): Message command type
- `sub_id` (optional): Sub-identifier (e.g., room ID, channel ID)
- `content` (required): Message content
- `metadata` (optional): Additional metadata as key-value pairs
- `priority` (optional): Message priority (1=high, 2=normal, 3=low, default=2). Used by consumers for handling order.

**Success Response (200 OK):**
```json
{
  "success": true,
  "message_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Message sent successfully",
  "data": {
    "queue_name": "message_queue",
    "priority": 2
  },
  "timestamp": 1234567890
}
```

**Error Response (400 Bad Request):**
```json
{
  "success": false,
  "error": "Validation failed: user_id is required",
  "code": "VALIDATION_ERROR",
  "timestamp": 1234567890
}
```

### GET /health

Health check endpoint for monitoring.

**Response (200 OK):**
```json
{
  "status": "healthy",
  "services": {
    "rabbitmq": true
  },
  "timestamp": 1234567890
}
```

**Response (503 Service Unavailable):**
```json
{
  "status": "unhealthy",
  "services": {
    "rabbitmq": false
  },
  "timestamp": 1234567890
}
```

### GET /

Root endpoint showing service information.

**Response (200 OK):**
```json
{
  "service": "RealTimeMessageChat REST API",
  "version": "1.0.0",
  "status": "running"
}
```

## Configuration

The application is configured via `config/api_server_config.json`:

### Server Configuration
- `host`: Server bind address (default: "0.0.0.0")
- `port`: Server port (default: 8080)
- `mode`: Gin mode - "debug", "release", or "test" (default: "release")
- `read_timeout_seconds`: HTTP read timeout (default: 30)
- `write_timeout_seconds`: HTTP write timeout (default: 30)
- `max_header_bytes`: Maximum header size in bytes (default: 1048576)
- `shutdown_timeout_seconds`: Graceful shutdown timeout (default: 15)

### RabbitMQ Configuration
- `host`: RabbitMQ server hostname (required)
- `port`: RabbitMQ server port (default: 5672)
- `username`: RabbitMQ username (default: "guest")
- `password`: RabbitMQ password (default: "guest")
- `vhost`: RabbitMQ virtual host (default: "/")
- `queue_name`: Queue name for publishing messages (required)
- `durable`: Queue durability (default: true)
- `auto_delete`: Auto-delete queue when unused (default: false)
- `exclusive`: Exclusive queue (default: false)
- `no_wait`: No-wait declaration (default: false)
- `connection_retry`: Connection retry attempts (1-5, default: 5)
- `retry_delay_seconds`: Delay between retries (default: 5)

### Logging Configuration
- `level`: Log level - "trace", "debug", "info", "warn", "error", "fatal", "panic" (default: "info")
- `format`: Log format - "json" or "text" (default: "json")
- `output_path`: Log file path (default: "logs/api_server.log", use "stdout" for console only)

## Usage Examples

### Using cURL

```bash
# Send a message
curl -X POST http://localhost:8080/api/v1/send_message \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "command": "chat_message",
    "content": "Hello, World!",
    "priority": 2
  }'

# Check health
curl http://localhost:8080/health
```

### Using HTTPie

```bash
# Send a message
http POST http://localhost:8080/api/v1/send_message \
  user_id=user123 \
  command=chat_message \
  content="Hello, World!" \
  priority:=2

# Check health
http GET http://localhost:8080/health
```

### Using Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func main() {
    message := map[string]interface{}{
        "user_id": "user123",
        "command": "chat_message",
        "content": "Hello, World!",
        "priority": 2,
    }

    jsonData, _ := json.Marshal(message)

    resp, err := http.Post(
        "http://localhost:8080/api/v1/send_message",
        "application/json",
        bytes.NewBuffer(jsonData),
    )

    // Handle response...
}
```

## Monitoring and Operations

### Logs

Logs are written to the path specified in the configuration:
- **Console**: Logs are always written to stdout
- **File**: Logs are also written to the specified file path
- **Format**: JSON format for easy parsing by log aggregation tools

### Health Checks

The `/health` endpoint can be used for:
- Load balancer health checks
- Container orchestration health probes
- Monitoring systems

### Metrics

For production deployments, consider integrating:
- Prometheus for metrics collection
- Grafana for visualization
- ELK stack for log aggregation

## Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s -X main.version=1.0.0" \
  -o api-server \
  ./cmd/server
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Vet code
go vet ./...
```

## Troubleshooting

### RabbitMQ Connection Issues

**Problem**: "Failed to connect to RabbitMQ"

**Solutions**:
1. Verify RabbitMQ is running: `docker ps | grep rabbitmq`
2. Check network connectivity: `telnet localhost 5672`
3. Verify credentials in configuration
4. Check RabbitMQ logs: `docker logs rabbitmq`

### Port Already in Use

**Problem**: "bind: address already in use"

**Solutions**:
1. Check what's using the port: `lsof -i :8080`
2. Change the port in configuration
3. Stop the conflicting service

### High Memory Usage

**Solutions**:
1. Adjust `GOMAXPROCS` environment variable
2. Review message throughput and consider scaling
3. Monitor with `pprof` for memory profiling

## Contributing

When contributing to the REST API:
1. Follow Go best practices and idioms
2. Maintain test coverage above 80%
3. Update documentation for API changes
4. Use structured logging for all operations
5. Handle errors gracefully with proper error codes

## License

This project is part of the RealTimeMessageChat system.

## Support

For issues, questions, or contributions, please refer to the main project repository.
