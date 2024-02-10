package routes

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/daffashafwan/tadarus-yuk/handlers"
	"github.com/daffashafwan/tadarus-yuk/internal/authorization"
)

// RegisterRoutes registers all the routes for the application.
func RegisterRoutes(router *mux.Router) {

	router.HandleFunc("/", handlers.Home).Methods(http.MethodGet)
	router.HandleFunc("/users/register", handlers.Register).Methods(http.MethodPost)
	router.HandleFunc("/users/login", handlers.Login).Methods(http.MethodPost)
	router.HandleFunc("/admin/login", handlers.Login).Methods(http.MethodPost)


	generalRoute := router.PathPrefix("/api").Subrouter()
	generalRoute.Use(authorization.AuthenticationMiddleware("user"))

	adminRoute := router.PathPrefix("/api").Subrouter()
	adminRoute.Use(authorization.AuthenticationMiddleware("admin"))

	// users
	generalRoute.HandleFunc("/users/{id}", handlers.GetUserByID).Methods(http.MethodGet)
	adminRoute.HandleFunc("/users", handlers.GetAllUsers).Methods(http.MethodGet)
	
	generalRoute.HandleFunc("/users/{id}", handlers.UpdateUser).Methods(http.MethodPut)
	adminRoute.HandleFunc("/users/{id}", handlers.DeleteUser).Methods(http.MethodDelete)

	// reading target
	generalRoute.HandleFunc("/users/{id}/reading-targets", handlers.CreateReadingTargetByUserID).Methods(http.MethodPost)
	generalRoute.HandleFunc("/users/{id}/reading-targets", handlers.GetAllReadingTargetByUserID).Methods(http.MethodGet)
	
	adminRoute.HandleFunc("/reading-targets", handlers.GetAllReadingTarget).Methods(http.MethodGet)
	generalRoute.HandleFunc("/reading-targets/{id}", handlers.GetReadingTargetByID).Methods(http.MethodGet)
	generalRoute.HandleFunc("/reading-targets/{id}", handlers.UpdateReadingTargetByID).Methods(http.MethodPut)
	generalRoute.HandleFunc("/reading-targets/{id}", handlers.DeleteReadingTarget).Methods(http.MethodDelete)

	// reading progress
	generalRoute.HandleFunc("/users/{id}/reading-progress", handlers.GetAllReadingProgressByUserID).Methods(http.MethodGet)
	generalRoute.HandleFunc("/users/{id}/reading-targets/{tid}/reading-progress", handlers.GetAllReadingProgressByUserIDTargetID).Methods(http.MethodGet)
	generalRoute.HandleFunc("/users/{id}/reading-targets/{tid}/reading-progress", handlers.CreateReadingProgress).Methods(http.MethodPost)

	adminRoute.HandleFunc("/reading-progress", handlers.GetAllReadingProgress).Methods(http.MethodGet)
	generalRoute.HandleFunc("/reading-progress/{id}", handlers.GetReadingProgressByID).Methods(http.MethodGet)
	generalRoute.HandleFunc("/reading-progress/{id}", handlers.UpdateReadingProgressByID).Methods(http.MethodPut)
	generalRoute.HandleFunc("/reading-progress/{id}", handlers.DeleteReadingProgress).Methods(http.MethodDelete)
	// Add more routes as needed
}