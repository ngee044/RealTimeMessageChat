package main

import (
	"Common/config"
	"RestAPI/handler"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Starting REST API Service...")

	config.ConnectRedis("RestAPI")
	config.ConnectRabbitMQ("RestAPI")

	r := mux.NewRouter()
	r.HandleFunc("/publish", handler.PublishMessageHandler).Methods("POST")

	port := config.GetEnv("PORT", "8080")
	log.Printf("Server started at :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
