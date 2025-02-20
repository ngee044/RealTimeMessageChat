#!/bin/bash

echo "Starting Redis Client..."
cd redis_client && go run main.go &

echo "Starting RabbitMQ Client..."
cd ../rabbitmq_client && go run main.go &

echo "Starting REST API Server..."
cd ../rest_api && go run main.go &

echo "All services are running."