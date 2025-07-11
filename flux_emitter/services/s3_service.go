package services

import (
	"context"
	"log"

	// "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SetupS3 initializes the AWS S3 client and configures the target bucket.
func SetupS3(bucketName string) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load AWS configuration: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	// Initialize the shared client and bucket name in flush_service
	InitS3(client, bucketName)
}
