package services

import (
	"context"
	"fmt"
	"strings"
	"encoding/json"
	// "log"
	

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var TopicBucket = "rayflux-store"  // You can move this to .env later

type ClientMetadata struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	LastUpdated string `json:"last_updated"`
	DeletedAt   string `json:"deleted_at,omitempty"`
	IsDeleted   bool   `json:"is_deleted"`
}

type TopicIndex struct {
	Topic string `json:"topic"`
	Publishers []ClientMetadata  `json:"publishers"`
	Subscribers []ClientMetadata  `json:"subscribers"`
}

// CreateTopic creates empty index files for publishers and subscribers for a topic
func CreateTopic(topic string) error {
	indexKey := fmt.Sprintf("rayflux/topic/%s.json", topic)

	index := TopicIndex{
		Topic:       topic,
		Publishers:  []ClientMetadata{},
		Subscribers: []ClientMetadata{},
	}

	indexBytes, err := json.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal topic index: %w", err)
	}

	_, err = S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &TopicBucket,
		Key:    &indexKey,
		Body:   strings.NewReader(string(indexBytes)),
	})

	return err
}


func ListTopics() ([]string, error) {
	prefix := "rayflux/topic/"

	// log.Printf("Listing topics with prefix: %s in bucket: %s", prefix, TopicBucket)

	resp, err := S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &TopicBucket,
		Prefix: &prefix,
	})
	if err !=nil{
		// log.Printf("Failed to list S3 objects: %v", err)
		return nil, err
	}

	// if len(resp.Contents) == 0 {
	// 	log.Println("No topics found in S3.")
	// }

	var topics []string
	for _, obj:= range resp.Contents{
		// log.Printf("Found object key: %s", *obj.Key)
		key := *obj.Key
		if strings.HasSuffix(key, ".json") {
			topic := strings.TrimSuffix(strings.TrimPrefix(key, prefix), ".json")
			// log.Printf("Parsed topic: %s", topic)
			topics = append(topics, topic)
		}
	}

	// log.Printf("Final topics list: %+v", topics)

	return topics, nil

}

func DeleteTopic(topic string) error {
	indexKey := fmt.Sprintf("rayflux/topic/%s.json", topic)

	_, err := S3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: &TopicBucket,
		Key:    &indexKey,
	})

	return err
}

func GetTopic(topic string) (*TopicIndex, error) {
	indexKey := fmt.Sprintf("rayflux/topic/%s.json", topic)

	resp, err := S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &TopicBucket,
		Key:    &indexKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch topic: %w", err)
	}
	defer resp.Body.Close()

	var index TopicIndex
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, fmt.Errorf("failed to parse topic index: %w", err)
	}

	return &index, nil
}