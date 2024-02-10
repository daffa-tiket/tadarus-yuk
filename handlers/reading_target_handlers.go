package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/daffashafwan/tadarus-yuk/db"
	"github.com/daffashafwan/tadarus-yuk/internal/dto"
	"github.com/gorilla/mux"
)

const (
	PagesAlQuran = 604
)

func GetAllReadingTarget(w http.ResponseWriter, r *http.Request) {
	// Query all reading_targets from the database
	query := "SELECT * FROM reading_target"
	rows, err := db.GetDB().Query(query)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching reading_targets"))
		return
	}
	defer rows.Close()

	var readingTargets []dto.ReadingTarget
	for rows.Next() {
		var readingTarget dto.ReadingTarget
		err := rows.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages, &readingTarget.Name, &readingTarget.StartPage, &readingTarget.EndPage)
		if err != nil {
			log.Printf("Error : %v", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error scanning reading_targets rows"))
			return
		}
		readingTargets = append(readingTargets, readingTarget)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readingTargets)
}

// GetReadingTargetByIDHandler handles requests to get a reading_target by ID.
func GetReadingTargetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingTargetID := vars["id"]

	readingTargetRes, err := getReadingTargetByID(readingTargetID)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching reading_target"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readingTargetRes)
}

func UpdateReadingTargetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingTargetID := vars["id"]

	readingTarget, err := getReadingTargetByID(readingTargetID)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching reading target"))
		return
	}

	var readingTargetUpdate dto.ReadingTarget
	err = json.NewDecoder(r.Body).Decode(&readingTargetUpdate)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}

	readingTarget.Name = readingTargetUpdate.Name
	readingTarget.StartDate = readingTargetUpdate.StartDate
	readingTarget.EndDate = readingTargetUpdate.EndDate
	readingTarget.Pages = readingTargetUpdate.Pages

	query := "UPDATE reading_target SET name = $1, start_date = $2, end_date = $3, start_page = $4, end_page = $5, target_pages_per_interval = $6 WHERE target_id = $7"
	_, err = db.GetDB().Exec(query, readingTarget.Name, readingTarget.StartDate, readingTarget.EndDate, readingTarget.StartPage, readingTarget.EndPage, readingTarget.Pages, readingTarget.ID)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error updating readingTarget data"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readingTarget)
}

func DeleteReadingTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingTargetID := vars["id"]

	// Delete the user from the database by ID
	query := "DELETE FROM reading_target WHERE target_id = $1"
	_, err := db.GetDB().Exec(query, readingTargetID)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error deleting reading target"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func CreateReadingTargetByUserID(w http.ResponseWriter, r *http.Request) {
	var readingTarget dto.ReadingTarget
	err := json.NewDecoder(r.Body).Decode(&readingTarget)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]

	_, err = getUserByID(userID)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error user not found"))
		return
	}

	if !isValidDateRange(readingTarget.StartDate, readingTarget.EndDate) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid date or date range"))
		return
	}

	if readingTarget.Pages < 1 || readingTarget.Pages > PagesAlQuran {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid number of pages"))
		return
	}

	readingTarget.UserID, _ = strconv.Atoi(userID)

	query := "INSERT INTO reading_target (user_id, name, start_date, end_date, start_page, end_page, target_pages_per_interval) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING target_id"
	err = db.GetDB().QueryRow(query, readingTarget.UserID, readingTarget.Name, readingTarget.StartDate, readingTarget.EndDate, readingTarget.StartPage, readingTarget.EndPage, readingTarget.Pages).Scan(&readingTarget.ID)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error creating reading target"))
		return
	}

	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(readingTarget)

}

func GetAllReadingTargetByUserID(w http.ResponseWriter, r *http.Request) {
	// Query all reading_targets from the database
	vars := mux.Vars(r)
	userID := vars["id"]

	query := "SELECT * FROM reading_target where user_id = $1"
	rows, err := db.GetDB().Query(query, userID)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching reading_targets"))
		return
	}
	defer rows.Close()

	var readingTargets []dto.ReadingTarget
	for rows.Next() {
		var readingTarget dto.ReadingTarget
		err := rows.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages, &readingTarget.Name, &readingTarget.StartPage, &readingTarget.EndPage)
		if err != nil {
			log.Printf("Error : %v", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error scanning reading_targets rows"))
			return
		}
		readingTargets = append(readingTargets, readingTarget)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readingTargets)
}

// getReadingTargetByID retrieves reading_target data from the database by ID.
func getReadingTargetByID(readingTargetID string) (dto.ReadingTarget, error) {
	// Query readingTarget data from the database by ID
	query := "SELECT * FROM reading_target WHERE target_id = $1"
	row := db.GetDB().QueryRow(query, readingTargetID)

	var readingTarget dto.ReadingTarget
	err := row.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages, &readingTarget.Name, &readingTarget.StartPage, &readingTarget.EndPage)
	if err == sql.ErrNoRows {
		return dto.ReadingTarget{}, fmt.Errorf("Reading Target with ID %s not found", readingTargetID)
	} else if err != nil {
		log.Printf("Error : %v", err.Error())
		return dto.ReadingTarget{}, err
	}

	return readingTarget, nil
}

func isValidDateRange(startDateStr, endDateStr string) bool {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		return false
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		return false
	}

	// Check if endDate is after startDate
	return endDate.After(startDate)
}
