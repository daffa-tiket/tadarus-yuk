package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/daffashafwan/tadarus-yuk/db"
	"github.com/daffashafwan/tadarus-yuk/internal/authorization"
	"github.com/daffashafwan/tadarus-yuk/internal/dto"
	"github.com/daffashafwan/tadarus-yuk/internal/helpers"
	"github.com/gorilla/mux"
)

// GetAllUsersHandler handles requests to get all users.
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Query all users from the database
	query := "SELECT * FROM users"
	rows, err := db.GetDB().Query(query)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching get all users", nil)
		return
	}
	defer rows.Close()

	var users []dto.User
	for rows.Next() {
		var user dto.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
		if err != nil {
			helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error scanning user row", nil)
			return
		}
		users = append(users, user)
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", users)
}

// GetUserByIDHandler handles requests to get a user by ID.
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	userResult, err := getUserByID(userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching get user by ID", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", userResult)
}

// RegisterHandler handles requests for user registration.
func Register(w http.ResponseWriter, r *http.Request) {
	var user dto.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Validate the password
	if err := helpers.ValidatePassword(user.Password); err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Hash the password
	hashedPassword, err := helpers.HashPassword(user.Password)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error hashing password", nil)
		return
	}

	// Insert the user into the database
	userResult, _ := getUserByUsername(user.Username)
	if userResult.Username == user.Username {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "username has taken", nil)
		return
	}
	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id"
	err = db.GetDB().QueryRow(query, user.Username, user.Email, hashedPassword).Scan(&user.ID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error creating user", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusCreated, "SUCCESS", user)
}

// UpdateUserHandler handles requests to update a user by ID.
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Fetch user data from the database by ID
	user, err := getUserByID(userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching user data", nil)
		return
	}

	// Decode the updated user data from the request body
	var updatedUser dto.User
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Invalid request body", nil)
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
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error updating user data", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", user)
}

// DeleteUserHandler handles requests to delete a user by ID.
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Delete the user from the database by ID
	query := "DELETE FROM users WHERE id = $1"
	_, err := db.GetDB().Exec(query, userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error deleting user", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusNoContent, "Succes Delete User", nil)
}

// LoginHandler handles requests for user login.
func Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest dto.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "invalid request body", nil)
		return
	}

	var userID int
	var authenticated bool
	var role string
	if strings.Contains(r.URL.Path, "admin") {
		authenticated, userID, err = authenticateAdmin(loginRequest.Username, loginRequest.Password)
		role = "admin"
	} else {
		authenticated, userID, err = authenticateUser(loginRequest.Username, loginRequest.Password)
		role = "user"
	}

	// Authenticate user (you may want to check the password against the hashed password in the database)
	// Example: Dummy authentication for illustration

	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "error authenticating user", nil)
		return
	}

	if !authenticated {
		helpers.ResponseJSON(w, err, http.StatusUnauthorized, "Invalid username or password", nil)
		return
	}

	// Generate authentication token (you may want to use a library like JWT)
	authToken, err := authorization.GenerateAuthToken(userID, role)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error generating authentication token", nil)
		return
	}

	// Return the authentication token
	resp := map[string]interface{} {
		"userID": userID,
		"token": authToken,
	}
	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", resp)
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
		log.Printf("Error : %v", err.Error())
		return dto.User{}, err
	}

	return user, nil
}

func authenticateUser(username, password string) (bool, int, error) {

	user, err := getUserByUsername(username)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		return false, 0, err
	}

	errVerify := helpers.VerifyPassword(password, user.Password)
	if errVerify != nil {
		return false, 0, nil
	}

	return true, user.ID, nil
}

func authenticateAdmin(username, password string) (bool, int, error) {

	admin, err := getAdminByUsername(username)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		return false, 0, err
	}

	errVerify := helpers.VerifyPassword(password, admin.Password)
	if errVerify != nil {
		return false, 0, nil
	}

	return true, admin.ID, nil
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
		log.Printf("Error : %v", err.Error())
		return dto.User{}, err
	}

	return user, nil
}

func getAdminByUsername(username string) (dto.Admin, error) {
	// Query user data from the database by username
	query := "SELECT * FROM admin WHERE username = $1"
	row := db.GetDB().QueryRow(query, username)

	var admin dto.Admin
	err := row.Scan(&admin.ID, &admin.Username, &admin.Email, &admin.Password)
	if err == sql.ErrNoRows {
		return dto.Admin{}, fmt.Errorf("username not found")
	} else if err != nil {
		log.Printf("Error : %v", err.Error())
		return dto.Admin{}, err
	}

	return admin, nil
}
