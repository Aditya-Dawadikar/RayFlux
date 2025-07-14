package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"flux_node/controllers"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/publish", controllers.PublishHandler).Methods("POST")
	r.HandleFunc("/subscribe", controllers.SubscribeHandler).Methods("GET")

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
