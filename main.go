package main

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/daffashafwan/tadarus-yuk/routes"
	"github.com/daffashafwan/tadarus-yuk/db"
	"github.com/daffashafwan/tadarus-yuk/env"
)

func main() {
	// Load environment variables
	env.LoadEnv()

	// Connect to the database
	db.ConnectDB()

	router := mux.NewRouter()

	routes.RegisterRoutes(router)

	port := ":9999"
	log.Printf("Server listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}