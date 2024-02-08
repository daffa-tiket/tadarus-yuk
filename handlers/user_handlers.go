package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/daffashafwan/tadarus-yuk/db"
	"github.com/daffashafwan/tadarus-yuk/internal/authorization"
	"github.com/daffashafwan/tadarus-yuk/internal/dto"
	"github.com/daffashafwan/tadarus-yuk/internal/helpers"
	"github.com/gorilla/mux"
)

// GetAllUsersHandler handles requests to get all users.
func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Query all users from the database
	query := "SELECT * FROM users"
	rows, err := db.GetDB().Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching users"))
		return
	}
	defer rows.Close()

	var users []dto.User
	for rows.Next() {
		var user dto.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error scanning user rows"))
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// GetUserByIDHandler handles requests to get a user by ID.
func GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	userResult, err := getUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching user"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResult)
}

// RegisterHandler handles requests for user registration.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user dto.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}

	// Validate the password
	if err := helpers.ValidatePassword(user.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Hash the password
	hashedPassword, err := helpers.HashPassword(user.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error hashing password"))
		return
	}

	// Insert the user into the database
	userResult, _ := getUserByUsername(user.Username)
	if userResult.Username == user.Username {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("username has taken"))
		return
	}
	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id"
	err = db.GetDB().QueryRow(query, user.Username, user.Email, hashedPassword).Scan(&user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error creating user"))
		return
	}

	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// UpdateUserHandler handles requests to update a user by ID.
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Fetch user data from the database by ID
	user, err := getUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching user data"))
		return
	}

	// Decode the updated user data from the request body
	var updatedUser dto.User
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}

	// Update user data based on the request body
	// For example, update user fields like username, email, etc.
	user.Username = updatedUser.Username
	user.Email = updatedUser.Email
	// Update other fields as needed

	// Save the updated user data to the database
	query := "UPDATE users SET username = $1, email = $2 WHERE id = $3"
	_, err = db.GetDB().Exec(query, user.Username, user.Email, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error updating user data"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// DeleteUserHandler handles requests to delete a user by ID.
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Delete the user from the database by ID
	query := "DELETE FROM users WHERE id = $1"
	_, err := db.GetDB().Exec(query, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error deleting user"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// LoginHandler handles requests for user login.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginRequest dto.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}

	// Authenticate user (you may want to check the password against the hashed password in the database)
	// Example: Dummy authentication for illustration
	authenticated, userID, err := authenticateUser(loginRequest.Username, loginRequest.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error authenticating user"))
		return
	}

	if !authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid username or password"))
		return
	}

	// Generate authentication token (you may want to use a library like JWT)
	authToken, err := authorization.GenerateAuthToken(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error generating authentication token"))
		return
	}

	// Return the authentication token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": authToken})
}

// getUserByID retrieves user data from the database by ID.
func getUserByID(userID string) (dto.User, error) {
	// Query user data from the database by ID
	query := "SELECT * FROM users WHERE id = $1"
	row := db.GetDB().QueryRow(query, userID)

	var user dto.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err == sql.ErrNoRows {
		return dto.User{}, fmt.Errorf("User with ID %s not found", userID)
	} else if err != nil {
		return dto.User{}, err
	}

	return user, nil
}

func authenticateUser(username, password string) (bool, int, error) {

	user, err := getUserByUsername(username)
	if err != nil {
		return false, 0, err
	}

	errVerify := helpers.VerifyPassword(password, user.Password)
	if errVerify != nil {
		return false, 0, nil
	}

	return true, user.ID, nil
}

// getUserByID retrieves user data from the database by username.
func getUserByUsername(username string) (dto.User, error) {
	// Query user data from the database by username
	query := "SELECT * FROM users WHERE username = $1"
	row := db.GetDB().QueryRow(query, username)

	var user dto.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err == sql.ErrNoRows {
		return dto.User{}, fmt.Errorf("username not found")
	} else if err != nil {
		return dto.User{}, err
	}

	return user, nil
}
