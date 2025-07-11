package main

import (
	"flux_emitter/controllers"
	"flux_emitter/services"
	"log"
	"os"
	"time"
	"strconv"

	"net/http"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Step 1: Initialize S3 Client
	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Fatal("S3_BUCKET not set in .env")
	}
	services.SetupS3(bucketName)

	// Step 2: Initialize Controllers
	bufferController := controllers.NewBufferController()


	// Step 3: Start Auto Flush every n seconds
	// Load flush interval from environment
	flushIntervalStr := os.Getenv("FLUSH_INTERVAL")
	flushInterval := 2 // Default: 1 seconds

	if flushIntervalStr != "" {
		if val, err := strconv.Atoi(flushIntervalStr); err == nil && val > 0 {
			flushInterval = val
		} else {
			log.Printf("Invalid FLUSH_INTERVAL '%s', using default %d seconds\n", flushIntervalStr, flushInterval)
		}
	} else {
		log.Printf("FLUSH_INTERVAL not set, using default %d seconds\n", flushInterval)
	}

	bufferController.StartAutoFlush(time.Duration(flushInterval) * time.Second)

	// Step 3: Setup Routes
	router := SetupRoutes(bufferController)

	// Step 4: Start HTTP Server
	log.Println("FluxEmitter server running on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
