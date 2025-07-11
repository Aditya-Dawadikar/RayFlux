package services

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	S3Client   *s3.Client
	BucketName string
)

func InitS3(client *s3.Client, bucket string) {
	S3Client = client
	BucketName = bucket
}

func FlushToS3(topic string, data []byte) error {
	if S3Client == nil{
		return fmt.Errorf("S3 client not initialized")
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000Z")
	key:= fmt.Sprintf("rayflux/%s/%s.jsonl", topic, timestamp)

	_, err := S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(BucketName),
		Key: aws.String(key),
		Body: bytes.NewReader(data),
	})

	if err!=nil{
		return fmt.Errorf("Failed to write to S3: %w", err)
	}

	return nil
}