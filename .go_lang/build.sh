#!/bin/bash
set -e  # 에러 발생 시 즉시 종료

echo "Starting build process..."

# Common 모듈 빌드
echo "Building Common module..."
cd Common
go mod tidy
go build ./...
cd ..

# RestAPI 빌드
echo "Setting up Common module in RestAPI..."
cd RestAPI
go mod edit -replace Common=../Common
go mod tidy
echo "Building RestAPI..."
go build ./...
cd ..

# MessageQueue 빌드
echo "Setting up Common module in MessageQueue..."
cd MessageQueue
go mod edit -replace Common=../Common
go mod tidy
echo "Building MessageQueue..."
go build ./...
cd ..

echo "Build process completed successfully!"
