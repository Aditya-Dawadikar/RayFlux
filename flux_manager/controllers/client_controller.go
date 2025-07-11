package controllers

import (
	"encoding/json"
	"flux_manager/services"
	"net/http"
)

type ClientController struct{}

func NewClientController() *ClientController {
	return &ClientController{}
}

// RegisterPublisher: POST /register-publisher
func (cc *ClientController) RegisterPublisher(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic      string `json:"topic"`
		PublisherID string `json:"publisher_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Topic == "" || req.PublisherID == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err := services.RegisterPublisher(req.Topic, req.PublisherID)
	if err != nil {
		http.Error(w, "Failed to register publisher: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"status":"publisher registered"}`))
}

// RegisterSubscriber: POST /register-subscriber
func (cc *ClientController) RegisterSubscriber(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic        string `json:"topic"`
		SubscriberID string `json:"subscriber_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Topic == "" || req.SubscriberID == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err := services.RegisterSubscriber(req.Topic, req.SubscriberID)
	if err != nil {
		http.Error(w, "Failed to register subscriber: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"status":"subscriber registered"}`))
}

// ListClients: GET /list-clients?topic=name
func (cc *ClientController) ListClients(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if topic == "" {
		http.Error(w, "Missing topic name", http.StatusBadRequest)
		return
	}

	data, err := services.GetTopic(topic)
	if err != nil {
		http.Error(w, "Failed to get topic: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(data)
}

// SoftDeletePublisher: DELETE /delete-publisher?topic=orders&publisher_id=pub-123
func (cc *ClientController) SoftDeletePublisher(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	publisherID := r.URL.Query().Get("publisher_id")

	if topic == "" || publisherID == "" {
		http.Error(w, "Missing topic or publisher_id", http.StatusBadRequest)
		return
	}

	err := services.SoftDeletePublisher(topic, publisherID)
	if err != nil {
		http.Error(w, "Failed to delete publisher: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"status":"publisher soft deleted"}`))
}

// SoftDeleteSubscriber: DELETE /delete-subscriber?topic=orders&subscriber_id=sub-456
func (cc *ClientController) SoftDeleteSubscriber(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	subscriberID := r.URL.Query().Get("subscriber_id")

	if topic == "" || subscriberID == "" {
		http.Error(w, "Missing topic or subscriber_id", http.StatusBadRequest)
		return
	}

	err := services.SoftDeleteSubscriber(topic, subscriberID)
	if err != nil {
		http.Error(w, "Failed to delete subscriber: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"status":"subscriber soft deleted"}`))
}
