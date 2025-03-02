package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"db_cli/cmd"
	"db_cli/internal/db"
	"db_cli/internal/server"
	"db_cli/internal/user"
)

func main() {
	godotenv.Load()

	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect DB: %v\n", err)
	}

	userService := user.NewUserService(database)

	cmd.RunCLI(userService)

	// run server mode
	if len(os.Args) > 1 && os.Args[1] == "--server" {
		port := os.Getenv("APP_PORT")
		if port == "" {
			port = "8080"
		}

		s := server.NewServer(userService)
		addr := fmt.Sprintf(":%s", port)
		log.Printf("Starting Gin server on %s\n", addr)
		if err := s.Run(addr); err != nil {
			log.Fatalf("Server error: %v\n", err)
		}
	}
}
