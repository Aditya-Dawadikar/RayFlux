package main

import (
	"flux_emitter/controllers"
	// "net/http"

	"github.com/gorilla/mux"
)

func SetupRoutes(bufferController *controllers.BufferController) *mux.Router{
	router := mux.NewRouter();

	publishController:= controllers.NewPublishController(bufferController)
	
	router.HandleFunc("/publish", publishController.HandlePublish).Methods("POST")

	return router
}