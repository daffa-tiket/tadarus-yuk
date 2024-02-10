package handlers

import "net/http"

// HomeHandler handles requests to the home endpoint.
func Home(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to the home page!"))
}
