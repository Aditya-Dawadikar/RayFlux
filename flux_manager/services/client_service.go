package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// RegisterPublisher adds a publisher to the topic index file
func RegisterPublisher(topic string, publisherID string) error {
	index, err := GetTopic(topic)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Check for existing publisher
	found := false
	for i, pub := range index.Publishers {
		if pub.ID == publisherID {
			// If deleted, undelete
			if pub.IsDeleted {
				index.Publishers[i].IsDeleted = false
				index.Publishers[i].DeletedAt = ""
			}
			index.Publishers[i].LastUpdated = now
			found = true
			break
		}
	}

	if !found {
		// Add new publisher
		newPub := ClientMetadata{
			ID:          publisherID,
			CreatedAt:   now,
			LastUpdated: now,
			IsDeleted:   false,
		}
		index.Publishers = append(index.Publishers, newPub)
	}

	return saveTopicIndex(topic, index)
}


// RegisterSubscriber adds a subscriber to the topic index file
func RegisterSubscriber(topic string, subscriberID string) error {
	index, err := GetTopic(topic)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Check for existing subscriber
	found := false
	for i, sub := range index.Subscribers {
		if sub.ID == subscriberID {
			if sub.IsDeleted {
				index.Subscribers[i].IsDeleted = false
				index.Subscribers[i].DeletedAt = ""
			}
			index.Subscribers[i].LastUpdated = now
			found = true
			break
		}
	}

	if !found {
		newSub := ClientMetadata{
			ID:          subscriberID,
			CreatedAt:   now,
			LastUpdated: now,
			IsDeleted:   false,
		}
		index.Subscribers = append(index.Subscribers, newSub)
	}

	return saveTopicIndex(topic, index)
}


// saveTopicIndex writes the updated topic index back to S3
func saveTopicIndex(topic string, index *TopicIndex) error {
	indexKey := fmt.Sprintf("rayflux/topic/%s.json", topic)

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

func SoftDeletePublisher(topic string, publisherID string) error {
	index, err := GetTopic(topic)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	found := false

	for i, pub := range index.Publishers {
		if pub.ID == publisherID && !pub.IsDeleted {
			index.Publishers[i].IsDeleted = true
			index.Publishers[i].DeletedAt = now
			index.Publishers[i].LastUpdated = now
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("publisher '%s' not found in topic '%s'", publisherID, topic)
	}

	return saveTopicIndex(topic, index)
}

func SoftDeleteSubscriber(topic string, subscriberID string) error {
	index, err := GetTopic(topic)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	found := false

	for i, sub := range index.Subscribers {
		if sub.ID == subscriberID && !sub.IsDeleted {
			index.Subscribers[i].IsDeleted = true
			index.Subscribers[i].DeletedAt = now
			index.Subscribers[i].LastUpdated = now
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("subscriber '%s' not found in topic '%s'", subscriberID, topic)
	}

	return saveTopicIndex(topic, index)
}
