package main

import (
	"Common/config"
	"MessageQueue/consumer"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	fmt.Println("Starting MessageQueue Consumer Service...")

	config.ConnectRedis("MessageQueue")
	config.ConnectRabbitMQ("MessageQueue")

	queueName := config.GetEnv("QUEUE_NAME", "task_queue")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		consumer.StartConsumer(queueName)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\nShutting down MessageQueue Consumer Service...")
		os.Exit(0)
	}()

	wg.Wait()
}
