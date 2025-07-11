package main

import (
	"flux_manager/services"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main(){
	err:= godotenv.Load()
	if err!=nil{
		log.Fatal("Error loading .env file")
	}

	// Step 1: Initialize S3 Client
	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Fatal("S3_BUCKET not set in .env")
	}
	services.SetupS3(bucketName)

	router := SetupRoutes()

	// Step 4: Start HTTP Server
	log.Println("FluxEmitter server running on :8081")
	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}