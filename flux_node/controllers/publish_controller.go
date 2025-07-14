package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"log"

	"github.com/gorilla/websocket"
)

type PublishRequest struct {
	Topic string `json:"topic"`
	Message string `json:"message"`
}

func PublishHandler(w http.ResponseWriter, r *http.Request){
	// Parse body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req PublishRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Topic == "" || req.Message == "" {
		http.Error(w, "Both 'topic' and 'message' are required", http.StatusBadRequest)
		return
	}

	// Initialize topic if needed
	InitTopic(req.Topic)

	msg := []byte(req.Message)

	// Convert to binary and store
	AddMessage(req.Topic, msg)

	// Step 2: fan-out to all subscribers
	subscriberMutex.RLock()
	subscribers, ok := Subscribers[req.Topic]
	subscriberMutex.RUnlock()

	if ok {
		for conn := range subscribers {
			go func(c *websocket.Conn) {
				if err := c.WriteMessage(websocket.BinaryMessage, msg); err != nil {
					log.Printf("[Fanout] Failed to write to subscriber: %v", err)
					removeSubscriber(req.Topic, c)
				}
			}(conn)
		}
	}

	// Respond
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"message accepted"}`))
}