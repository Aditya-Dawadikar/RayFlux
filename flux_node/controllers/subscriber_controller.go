package controllers

import (
	"log"
	"net/http"
	"sync"
	"encoding/json"

	"github.com/gorilla/websocket"
)

var Subscribers = make(map[string]map[*websocket.Conn]bool)
var subscriberMutex sync.RWMutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	// Step 1: read initial subscribe message (JSON with topic + subscriber_id)
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("[SubscribeHandler] Failed to read subscribe message:", err)
		conn.Close()
		return
	}

	var req struct {
		SubscriberID string `json:"subscriber_id"`
		Topic        string `json:"topic"`
	}
	if err := json.Unmarshal(msg, &req); err != nil || req.Topic == "" || req.SubscriberID == "" {
		log.Printf("[SubscribeHandler] Invalid subscribe payload: %s", string(msg))
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid subscribe request"}`))
		conn.Close()
		return
	}

	InitTopic(req.Topic)

	// Register the connection
	subscriberMutex.Lock()
	if _, ok := Subscribers[req.Topic]; !ok {
		Subscribers[req.Topic] = make(map[*websocket.Conn]bool)
	}
	Subscribers[req.Topic][conn] = true
	subscriberMutex.Unlock()

	log.Printf("[Subscriber] Connected: %s on topic '%s'", req.SubscriberID, req.Topic)

	// Wait for disconnect
	go waitForDisconnect(req.Topic, conn)
}


func waitForDisconnect(topic string, conn *websocket.Conn) {
	for {
		_, _, err := conn.NextReader()
		if err != nil {
			log.Printf("[Subscriber] Disconnected from topic '%s'", topic)
			removeSubscriber(topic, conn)
			return
		}
	}
}

func removeSubscriber(topic string, conn *websocket.Conn) {
	subscriberMutex.Lock()
	defer subscriberMutex.Unlock()

	if conns, ok := Subscribers[topic]; ok {
		delete(conns, conn)
		conn.Close()

		if len(conns) == 0 {
			delete(Subscribers, topic)
		}
	}
}
