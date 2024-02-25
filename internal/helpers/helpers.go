package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/daffashafwan/tadarus-yuk/internal/dto"
	"golang.org/x/crypto/bcrypt"
)

// validatePassword checks if the password meets your criteria (e.g., length).
func ValidatePassword(password string) error {
	// Implement your password validation logic here
	// For example, check the length or any other criteria
	if len(password) < 8 {
		return errors.New("Password must be at least 8 characters long")
	}
	return nil
}

// hashPassword hashes the given password using bcrypt.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		return "", err
	}
	return string(hashedPassword), nil
}

func VerifyPassword(providedPassword, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
}

func ResponseJSON(w http.ResponseWriter, err error, statusCode int, message string, data interface{}) {
	messages := []string{
		message,
	}
	if err != nil {
		log.Printf("Error : %v", err.Error())
		messages = append(messages, err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	dataRes := dto.Response{
		Code:    statusCode,
		Message: messages,
		Data:    data,
	}
	resp, _ := json.Marshal(dataRes)
	w.Write(resp)
}

func BuildInClause(listID []int) string {
    var inClause string
    for i, id := range listID {
        inClause += fmt.Sprintf("%d", id)
        if i < len(listID)-1 {
            inClause += ", "
        }
    }
    return inClause
}
