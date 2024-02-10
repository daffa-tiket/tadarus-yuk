package routes

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/daffashafwan/tadarus-yuk/handlers"
)

// RegisterRoutes registers all the routes for the application.
func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/", handlers.Home).Methods(http.MethodGet)

	// users
	router.HandleFunc("/api/users/{id}", handlers.GetUserByID).Methods(http.MethodGet)
	router.HandleFunc("/api/users", handlers.GetAllUsers).Methods(http.MethodGet)
	router.HandleFunc("/api/users", handlers.Register).Methods(http.MethodPost)
	router.HandleFunc("/api/users/{id}", handlers.UpdateUser).Methods(http.MethodPut)
	router.HandleFunc("/api/users/{id}", handlers.DeleteUser).Methods(http.MethodDelete)

	// auth
	router.HandleFunc("/api/users/login", handlers.Login).Methods(http.MethodPost)

	// reading target
	router.HandleFunc("/api/users/{id}/reading-targets", handlers.CreateReadingTargetByUserID).Methods(http.MethodPost)
	router.HandleFunc("/api/users/{id}/reading-targets", handlers.GetAllReadingTargetByUserID).Methods(http.MethodGet)
	router.HandleFunc("/api/reading-targets", handlers.GetAllReadingTarget).Methods(http.MethodGet)
	router.HandleFunc("/api/reading-targets/{id}", handlers.GetReadingTargetByID).Methods(http.MethodGet)
	router.HandleFunc("/api/reading-targets/{id}", handlers.UpdateReadingTargetByID).Methods(http.MethodPut)
	router.HandleFunc("/api/reading-targets/{id}", handlers.DeleteReadingTarget).Methods(http.MethodDelete)
	// Add more routes as needed
}