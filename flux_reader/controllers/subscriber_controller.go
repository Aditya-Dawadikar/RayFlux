package controllers

import (
	"encoding/json"
	"flux_reader/services"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
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

	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Failed to read subscription request: %v", err)
		return
	}

	var req SubscribeRequest
	if err := json.Unmarshal(message, &req); err != nil || req.Topic == "" || req.SubscriberID == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Invalid subscription request"}`))
		return
	}

	log.Printf("Subscriber [%s] for topic [%s] connected.", req.SubscriberID, req.Topic)

	
	doneChan := make(chan struct{})
	ackChan := make(chan map[string]string, 1)

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("ReadMessage error: %v", err)
				close(doneChan)
				return
			}

			// Try to decode ACK
			var ack map[string]string
			if json.Unmarshal(msg, &ack) == nil && ack["batch_id"] != "" {
				select {
				case ackChan <- ack:
				default:
					// drop if nobody is waiting
				}
			}
		}
	}()

	// WebSocket ping-pong for dead connection detection
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	conn.SetPongHandler(func(appData string) error {
		log.Printf("Received Pong from client: %s", appData)
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		return nil
	})


	go services.StartPingLoop(conn, doneChan)

	checkpoint, err := services.LoadCheckpoint(req.Topic, req.SubscriberID)
	if err != nil {
		log.Printf("Failed to load checkpoint: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Failed to load checkpoint"}`))
		return
	}

	services.PollAndStreamMessages(conn, req.Topic, req.SubscriberID, checkpoint, doneChan, ackChan)

	log.Printf("Connection closed for subscriber [%s] on topic [%s]", req.SubscriberID, req.Topic)
}

