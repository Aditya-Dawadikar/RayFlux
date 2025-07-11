package controllers

import (
	"encoding/json"
	"net/http"
)

type PublishRequest struct {
	Topic string `json: "topic"`
	Message string `json: "message"`
}

type PublishController struct {
	Buffer *BufferController
}


func NewPublishController(buffer *BufferController) *PublishController{
	return &PublishController{
		Buffer: buffer,
	}
}

func (pc *PublishController) HandlePublish(w http.ResponseWriter, r *http.Request){
	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Topic == "" || req.Message == "" {
		http.Error(w, "Both topic and message are required", http.StatusBadRequest)
		return
	}

	// add message to cache
	err := pc.Buffer.AddMessage(req.Topic, []byte(req.Message))
	if err != nil{
		http.Error(w, "Failed to process message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}