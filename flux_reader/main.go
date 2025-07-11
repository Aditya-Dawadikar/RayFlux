package main

import (
	"flag"
	"flux_reader/services"
	"flux_reader/config"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	flag.StringVar(&config.ReaderID, "reader", "flux-reader-default", "Name of this FluxReader")
	flag.Parse()

	// Step 1: Load environment variables
	env := os.Getenv("ENV")
	if env == "dev_local" {
		err := godotenv.Load()
		if err != nil {
			log.Println("No .env file found, using system environment")
		} else {
			log.Println(".env file loaded successfully")
		}
	} else {
		log.Println("Running in", env, "— using environment variables only")
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
