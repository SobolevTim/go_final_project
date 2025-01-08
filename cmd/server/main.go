package main

import (
	"log"

	"GO_FINAL_PROJECT/config"
	"GO_FINAL_PROJECT/internal/database"
	"GO_FINAL_PROJECT/internal/server"
)

func main() {
	// Load the config from environment variables
	cfg := config.LoadCongig()

	// Connect to the database
	service, err := database.Conect(cfg.DBFile)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer service.DB.Close()
	// Start the server
	if err := server.StartServer(cfg, service); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
