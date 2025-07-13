package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"

	"flux_balancer/internal/discovery"
	"flux_balancer/internal/hash"
)

type SubscribeRequest struct {
	SubscriberID string `json:"subscriber_id"`
	Topic        string `json:"topic"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // loosen for dev
}

func ProxySubscribe(w http.ResponseWriter, r *http.Request) {
	// Step 1: Upgrade client HTTP connection to WebSocket
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ProxySubscribe] Failed to upgrade WebSocket: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}
	defer clientConn.Close()

	// Step 2: Read initial subscribe request from client
	_, msg, err := clientConn.ReadMessage()
	if err != nil {
		log.Printf("[ProxySubscribe] Failed to read subscription request: %v", err)
		return
	}

	var req SubscribeRequest
	if err := json.Unmarshal(msg, &req); err != nil || req.SubscriberID == "" || req.Topic == "" {
		log.Printf("[ProxySubscribe] Invalid subscription payload: %s", string(msg))
		clientConn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Invalid subscription request"}`))
		return
	}
	log.Printf("[ProxySubscribe] Subscriber=%s, Topic=%s", req.SubscriberID, req.Topic)

	// Step 3: Choose FluxReader pod
	podIPs := discovery.GetReaderAddrs()
	if len(podIPs) == 0 {
		clientConn.WriteMessage(websocket.TextMessage, []byte(`{"error":"No reader pods available"}`))
		return
	}
	index := hash.GetConsistentIndex(req.SubscriberID, req.Topic, podIPs)
	if index == -1 {
		clientConn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Hashing failed"}`))
		return
	}
	target := fmt.Sprintf("ws://%s:8082/subscribe", podIPs[index])
	log.Printf("[ProxySubscribe] Routing to reader pod: %s", target)

	// Step 4: Connect to the chosen reader pod
	readerURL := url.URL{Scheme: "ws", Host: podIPs[index] + ":8082", Path: "/subscribe"}
	readerConn, _, err := websocket.DefaultDialer.Dial(readerURL.String(), nil)
	if err != nil {
		log.Printf("[ProxySubscribe] Failed to connect to reader pod: %v", err)
		clientConn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Failed to connect to reader pod"}`))
		return
	}
	defer readerConn.Close()

	// Step 5: Forward initial message to reader
	if err := readerConn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Printf("[ProxySubscribe] Failed to forward subscription request: %v", err)
		return
	}

	// Step 6: Bidirectional proxying
	errc := make(chan error, 2)

	go proxyWebSocket(clientConn, readerConn, errc)
	go proxyWebSocket(readerConn, clientConn, errc)

	<-errc // wait for either direction to fail
}

func proxyWebSocket(src, dst *websocket.Conn, errc chan error) {
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			errc <- err
			return
		}
		if err := dst.WriteMessage(msgType, msg); err != nil {
			errc <- err
			return
		}
	}
}
