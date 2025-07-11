package main

import (
	"flux_reader/controllers"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// WebSocket endpoint for subscribers
	router.HandleFunc("/subscribe", controllers.HandleWebSocket)

	return router
}
