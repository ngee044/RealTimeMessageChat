#!/bin/bash
set -e  # 에러 발생 시 즉시 종료

echo "Starting build process..."

# Common 모듈 빌드
echo "Building Common module..."
cd Common || { echo "Error: Cannot access Common directory"; exit 1; }
go mod tidy
go build ./...
cd ..

# RestAPI 빌드
echo "Setting up Common module in RestAPI..."
cd RestAPI || { echo "Error: Cannot access RestAPI directory"; exit 1; }
go mod edit -replace Common=../Common
go mod tidy
echo "Building RestAPI..."
go build ./... || { echo "Error: Failed to build RestAPI"; exit 1; }
cd ..

# MessageQueue 빌드
echo "Setting up Common module in MessageQueue..."
cd MessageQueue || { echo "Error: Cannot access MessageQueue directory"; exit 1; }
go mod edit -replace Common=../Common
go mod tidy
echo "Building MessageQueue..."
go build ./... || { echo "Error: Failed to build MessageQueue"; exit 1; }
cd ..

# DBClient 빌드
echo "Setting up Common module in DBClient..."
cd db_cli || { echo "Error: Cannot access DBClient directory"; exit 1; }
go mod edit -replace Common=../Common
go mod tidy
echo "Building DBClient..."
go build ./... || { echo "Error: Failed to build DBClient"; exit 1; }
cd ..

echo "Build process completed successfully!"
