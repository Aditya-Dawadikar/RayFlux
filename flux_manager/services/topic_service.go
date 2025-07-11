package services

import (
	"context"
	"fmt"
	"strings"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)


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


	resp, err := S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &TopicBucket,
		Prefix: &prefix,
	})
	if err !=nil{
		return nil, err
	}

	var topics []string
	for _, obj:= range resp.Contents{
		key := *obj.Key
		if strings.HasSuffix(key, ".json") {
			topic := strings.TrimSuffix(strings.TrimPrefix(key, prefix), ".json")
			topics = append(topics, topic)
		}
	}


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