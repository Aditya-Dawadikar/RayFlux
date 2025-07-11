package main

import (
	"flux_manager/services"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main(){
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

	// Step 1: Initialize S3 Client
	services.SetupS3()

	router := SetupRoutes()

	// Step 4: Start HTTP Server
	log.Println("FluxEmitter server running on :8081")
	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}