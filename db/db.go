package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

// ConnectDB connects to the PostgreSQL database.
func ConnectDB() {
	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		log.Fatal("Error connecting to the database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Printf("Error : %v", err.Error())
		log.Fatal("Error pinging the database:", err)
	}

	fmt.Println("Connected to the database")
}

// GetDB returns the database connection.
func GetDB() *sql.DB {
	return db
}
