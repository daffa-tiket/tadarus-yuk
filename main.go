package main

import (
	"log"
	"net/http"
	"os"

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

	useTLS := os.Getenv("USE_TLS")

	routes.RegisterRoutes(router)

	log.Print("Server starting\n")

	var err error
	if useTLS == "true" {
		// Specify the paths to your SSL certificate and private key files
		certFile := os.Getenv("CERT_FILE")
	    keyFile := os.Getenv("KEY_FILE")
		err = http.ListenAndServeTLS(":443", certFile, keyFile, handlers.CORS(headersOk, originsOk, methodsOk)(router))
	} else {
		err = http.ListenAndServe(":80", handlers.CORS(headersOk, originsOk, methodsOk)(router))
	}

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}