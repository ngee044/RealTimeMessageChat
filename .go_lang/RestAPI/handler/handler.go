package handler

import (
	"Common/config"
	"fmt"
	"log"
	"net/http"
)

// PublishMessageHandler handles RabbitMQ message publishing
func PublishMessageHandler(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Query().Get("message")

	err := config.PublishMessage(config.GetEnv("QUEUE_NAME", "task_queue"), message)
	if err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	log.Printf("Published: %s", message)
	fmt.Fprintf(w, "Message published: %s", message)
}
