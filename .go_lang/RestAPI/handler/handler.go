package handler

import (
	"Common/config"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// PublishMessageHandler handles RabbitMQ message publishing
func PublishMessageHandler(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Query().Get("message")
	if message == "" {
		http.Error(w, "Message parameter is required", http.StatusBadRequest)
		return
	}

	err := config.PublishMessage(config.GetEnv("QUEUE_NAME", "task_queue"), message)
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	log.Printf("Published: %s", message)
	fmt.Fprintf(w, "Message published: %s", message)
}

// PollMessageHandler handles polling requests from clients
func PollMessageHandler(w http.ResponseWriter, r *http.Request) {
	message, err := config.GetKey("latest_message")
	if err != nil {
		log.Printf("Failed to retrieve message from Redis: %v", err)
		http.Error(w, "Failed to retrieve message", http.StatusInternalServerError)
		return
	}

	if message == "" {
		http.Error(w, "No messages found", http.StatusNotFound)
		return
	}

	response := map[string]string{"message": message}
	json.NewEncoder(w).Encode(response)
}
