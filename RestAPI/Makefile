.PHONY: help build run test clean docker-build docker-run fmt lint vet tidy swagger swagger-validate

# Variables
BINARY_NAME=api-server
DOCKER_IMAGE=rtmc-api-server
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/server
	@echo "Build complete: $(BINARY_NAME)"

run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	@go run ./cmd/server -config config/api_server_config.json

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Tests complete"

test-coverage: test ## Run tests with coverage report
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.txt coverage.html
	@rm -rf logs/
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(VERSION)..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@echo "Docker image built: $(DOCKER_IMAGE):$(VERSION)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm -p 8080:8080 \
		-v $(PWD)/config:/app/config \
		-v $(PWD)/logs:/app/logs \
		$(DOCKER_IMAGE):latest

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

lint: ## Lint Go code
	@echo "Linting code..."
	@golangci-lint run
	@echo "Lint complete"

vet: ## Vet Go code
	@echo "Vetting code..."
	@go vet ./...
	@echo "Vet complete"

tidy: ## Tidy Go modules
	@echo "Tidying modules..."
	@go mod tidy
	@echo "Tidy complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Tools installed"

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger docs..."
	@swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	@echo "Fixing swag version compatibility..."
	@sed -i '' '/LeftDelim:/d' docs/docs.go 2>/dev/null || true
	@sed -i '' '/RightDelim:/d' docs/docs.go 2>/dev/null || true
	@echo "Swagger docs generated"

swagger-validate: ## Validate and format Swagger comments
	@echo "Validating Swagger comments..."
	@swag fmt -g cmd/server/main.go
	@echo "Swagger validation complete"

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

build-release: ## Build release binary for multiple platforms
	@echo "Building release binaries..."
	@mkdir -p build
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-linux-amd64 ./cmd/server
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-linux-arm64 ./cmd/server
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-darwin-amd64 ./cmd/server
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o build/$(BINARY_NAME)-darwin-arm64 ./cmd/server
	@echo "Release binaries built in build/"

dev: ## Run in development mode with auto-reload (requires air)
	@air

.DEFAULT_GOAL := help
