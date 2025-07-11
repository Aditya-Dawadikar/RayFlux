package main

import (
	"flux_reader/services"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Step 1: Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default AWS environment")
	}

	// Step 2: Initialize S3 Client and Bucket
	services.SetupS3()

	// Step 3: Setup routes
	router := SetupRoutes()

	// Step 4: Start WebSocket Server
	port := os.Getenv("READER_PORT")
	if port == "" {
		port = "8082" // Default port if not set
	}

	log.Printf("FluxReader running on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
