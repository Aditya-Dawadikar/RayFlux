package main

import (
	"flux_manager/controllers"
	// "net/http"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Topic Controller
	topicController := controllers.NewTopicController()

	router.HandleFunc("/topic", topicController.CreateTopic).Methods("POST")
	router.HandleFunc("/topic", topicController.GetTopic).Methods("GET")
	router.HandleFunc("/topic", topicController.DeleteTopic).Methods("DELETE")
	router.HandleFunc("/topics", topicController.ListTopics).Methods("GET")

	// Client Controller
	clientController := controllers.NewClientController()
	router.HandleFunc("/publisher", clientController.RegisterPublisher).Methods("POST")
	router.HandleFunc("/publisher", clientController.SoftDeletePublisher).Methods("DELETE")
	router.HandleFunc("/subscriber", clientController.RegisterSubscriber).Methods("POST")
	router.HandleFunc("/subscriber", clientController.SoftDeleteSubscriber).Methods("DELETE")
	router.HandleFunc("/clients", clientController.ListClients).Methods("GET")

	return router
}
