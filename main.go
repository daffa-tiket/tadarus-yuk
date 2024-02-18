package main

import (
	"log"
	"net/http"

	"github.com/daffashafwan/tadarus-yuk/db"
	"github.com/daffashafwan/tadarus-yuk/env"
	"github.com/daffashafwan/tadarus-yuk/external"
	"github.com/daffashafwan/tadarus-yuk/internal/authorization"
	"github.com/daffashafwan/tadarus-yuk/routes"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Load environment variables
	env.LoadEnv()

	// Connect to the database
	db.ConnectDB()

	external.InitQuranAPI()

	authorization.InitSecret()

	router := mux.NewRouter()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	routes.RegisterRoutes(router)

	port := ":9999"
	log.Printf("Server listening on port %s...\n", port)

	log.Fatal(http.ListenAndServe(port, handlers.CORS(headersOk, originsOk, methodsOk)(router)))
}