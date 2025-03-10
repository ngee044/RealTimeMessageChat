package config

import (
	"log"

	"github.com/streadway/amqp"
)

var rabbitConn *amqp.Connection
var rabbitChannel *amqp.Channel

// ConnectRabbitMQ initializes RabbitMQ connection
func ConnectRabbitMQ(service string) {
	LoadEnv(service) // 서비스별 환경 변수 로드

	var err error
	rabbitConn, err = amqp.Dial(GetEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	rabbitChannel, err = rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	log.Println("Connected to RabbitMQ")
}

// PublishMessage sends a message to RabbitMQ
func PublishMessage(queueName, message string) error {
	_, err := rabbitChannel.QueueDeclare(
		queueName,
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return rabbitChannel.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}
