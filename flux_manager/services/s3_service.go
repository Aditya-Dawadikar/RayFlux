package services

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3Client *s3.Client
var TopicBucket string

func SetupS3() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	S3Client = s3.NewFromConfig(cfg)

	TopicBucket = os.Getenv("S3_BUCKET")
	if TopicBucket == "" {
		log.Fatal("S3_BUCKET not set in environment")
	}
}
