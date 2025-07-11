package services

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// CleanupConnection stops heartbeat, stops polling, and ends session in FluxManager
func CleanupConnection(managerURL, topic, subscriberID, readerID string, stopChan chan struct{}) {
	// Step 1: Stop all background goroutines
	if stopChan != nil {
		close(stopChan)
	}

	// Step 2: Call FluxManager to end session
	payload := map[string]string{
		"topic":         topic,
		"subscriber_id": subscriberID,
		"reader_id":     readerID,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(managerURL+"/session/end", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("❌ Failed to notify session end for [%s:%s]: %v", topic, subscriberID, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Session closed for [%s:%s]", topic, subscriberID)
}
