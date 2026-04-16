package main

import (
	"context"
	"fmt"
	"mangascraper/internal/db"
	"mangascraper/internal/handler"
	"mangascraper/internal/repository"
	"mangascraper/internal/service"
	"os"
)

func main() {
	ctx := context.Background()

	if err := db.InitDB(ctx); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	defer db.CloseDB()

	queries := repository.New(db.Pool)

	if err := handler.InitScraper(); err != nil {
		fmt.Printf("Warning: Failed to initialize scraper: %v\n", err)
	}

	downloadSvc := service.NewDownloadService(queries)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	server := handler.NewServer(port, queries, downloadSvc)
	if err := server.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
