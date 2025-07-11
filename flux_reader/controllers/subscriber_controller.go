package controllers

import (
	"encoding/json"
	"flux_reader/services"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

type SubscribeRequest struct {
	Topic        string `json:"topic"`
	SubscriberID string `json:"subscriber_id"`
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Println("Subscriber connected.")

	// Step 1: Receive initial message → extract topic and subscriber_id
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Failed to read initial message: %v", err)
		return
	}

	var req SubscribeRequest
	if err := json.Unmarshal(message, &req); err != nil || req.Topic == "" || req.SubscriberID == "" {
		log.Printf("Invalid subscription request: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Invalid subscription request"}`))
		return
	}

	log.Printf("Subscriber [%s] for topic [%s] connected.", req.SubscriberID, req.Topic)

	// Step 2: Load checkpoint
	checkpoint, err := services.LoadCheckpoint(req.Topic, req.SubscriberID)
	if err != nil {
		log.Printf("Failed to load checkpoint: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Failed to load checkpoint"}`))
		return
	}

	// Step 3: Start polling and streaming loop (blocking)
	services.PollAndStreamMessages(conn, req.Topic, req.SubscriberID, checkpoint)
}
