package main

import (
	"fmt"
	"mangascraper/internal/api"
	"mangascraper/internal/db"
	"os"
)

func main() {
	// Initialize database
	if err := db.InitDB(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	defer db.CloseDB()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create and start HTTP server
	server := api.NewServer(port)
	if err := server.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
