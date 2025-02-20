package consumer

import (
	"Common/config"
	"log"

	"github.com/streadway/amqp"
)

// StartConsumer listens for messages from RabbitMQ and stores in Redis
func StartConsumer(queueName string) {
	conn, err := amqp.Dial(config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	go func() {
		for msg := range msgs {
			log.Printf(" [x] Received: %s", string(msg.Body))

			// 받은 메시지를 Redis에 저장
			err := config.SetKey("latest_message", string(msg.Body))
			if err != nil {
				log.Printf("Failed to store message in Redis: %s", err)
			} else {
				log.Println("Message stored in Redis")
			}
		}
	}()

	log.Println("Waiting for messages...")
	select {} // 서비스 유지
}
