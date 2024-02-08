package helpers

import (
	"errors"
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
		return "", err
	}
	return string(hashedPassword), nil
}

func VerifyPassword(providedPassword, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
}
