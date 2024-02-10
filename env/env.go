package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file.
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error : %v", err.Error())
		log.Fatal("Error loading .env file")
	}
}

// GetEnvVar retrieves the value of an environment variable by name.
func GetEnvVar(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}
