package routes

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/daffashafwan/tadarus-yuk/handlers"
)

// RegisterRoutes registers all the routes for the application.
func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/", handlers.HomeHandler).Methods(http.MethodGet)

	// users
	router.HandleFunc("/api/users/{id}", handlers.GetUserByIDHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/users", handlers.GetAllUsersHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/users", handlers.RegisterHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/users/{id}", handlers.UpdateUserHandler).Methods(http.MethodPut)
	router.HandleFunc("/api/users/{id}", handlers.DeleteUserHandler).Methods(http.MethodDelete)

	//auth
	router.HandleFunc("/api/users/login", handlers.LoginHandler).Methods(http.MethodPost)
	// Add more routes as needed
}