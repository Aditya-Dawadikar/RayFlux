package controllers

import (
	"encoding/json"
	"flux_manager/services"
	"net/http"
)

type TopicController struct{}

func NewTopicController() *TopicController {
	return &TopicController{}
}

// CreateTopic: POST /create-topic
func (tc *TopicController) CreateTopic(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic string `json:"topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Topic == "" {
		http.Error(w, "Invalid topic name", http.StatusBadRequest)
		return
	}

	err := services.CreateTopic(req.Topic)
	if err != nil {
		http.Error(w, "Failed to create topic: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"topic created"}`))
}

// ListTopics: GET /list-topics
func (tc *TopicController) ListTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := services.ListTopics()
	if err != nil {
		http.Error(w, "Failed to list topics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"topics": topics,
	})
}

// GetTopic: GET /topic
func (tc *TopicController) GetTopic(w http.ResponseWriter, r *http.Request) {
	topicName := r.URL.Query().Get("name")
	if topicName == "" {
		http.Error(w, "Missing topic name in query params", http.StatusBadRequest)
		return
	}

	topic, err := services.GetTopic(topicName)
	if err != nil {
		http.Error(w, "Failed to get topic: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(topic)
}


// DeleteTopic: DELETE /delete-topic
func (tc *TopicController) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	topicName := r.URL.Query().Get("name")
	if topicName == "" {
		http.Error(w, "Missing topic name in query params", http.StatusBadRequest)
		return
	}

	err := services.DeleteTopic(topicName)
	if err != nil {
		http.Error(w, "Failed to delete topic: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"topic deleted"}`))
}
