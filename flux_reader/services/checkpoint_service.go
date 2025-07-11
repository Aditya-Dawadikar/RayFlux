package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type SubscriberCheckpoint struct {
	Topic           string `json:"topic"`
	SubscriberID    string `json:"subscriber_id"`
	LastReadFile    string `json:"last_read_file"`
	LastReadTimeUTC string `json:"last_read_time_utc"`
}

// LoadCheckpoint loads the checkpoint for a subscriber from S3
func LoadCheckpoint(topic string, subscriberID string) (*SubscriberCheckpoint, error) {
	checkpointKey := fmt.Sprintf("rayflux/checkpoints/%s/%s.json", topic, subscriberID)

	resp, err := S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &BucketName,
		Key:    &checkpointKey,
	})
	if err != nil {
		// If no checkpoint found → return empty starting point
		if strings.Contains(err.Error(), "NoSuchKey") {
			return &SubscriberCheckpoint{
				Topic:           topic,
				SubscriberID:    subscriberID,
				LastReadFile:    "",
				LastReadTimeUTC: time.Now().UTC().Format(time.RFC3339),
			}, nil
		}
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}
	defer resp.Body.Close()

	var checkpoint SubscriberCheckpoint
	if err := json.NewDecoder(resp.Body).Decode(&checkpoint); err != nil {
		return nil, fmt.Errorf("failed to decode checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// SaveCheckpoint writes the checkpoint back to S3
func SaveCheckpoint(topic string, subscriberID string, checkpoint *SubscriberCheckpoint) error {
	checkpointKey := fmt.Sprintf("rayflux/checkpoints/%s/%s.json", topic, subscriberID)

	checkpoint.LastReadTimeUTC = time.Now().UTC().Format(time.RFC3339)

	data, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	_, err = S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &BucketName,
		Key:    &checkpointKey,
		Body:   strings.NewReader(string(data)),
	})

	return err
}
