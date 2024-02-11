package main

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
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

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	routes.RegisterRoutes(router)

	port := ":9999"
	log.Printf("Server listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, router))

	log.Fatal(http.ListenAndServe(port, handlers.CORS(headersOk, originsOk, methodsOk)(router)))
}