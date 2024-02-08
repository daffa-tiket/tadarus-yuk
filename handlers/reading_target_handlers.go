package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/daffashafwan/tadarus-yuk/db"
	"github.com/daffashafwan/tadarus-yuk/internal/dto"
	"github.com/gorilla/mux"
)

func GetAllReadingTargetHandler(w http.ResponseWriter, r *http.Request) {
	// Query all reading_targets from the database
	query := "SELECT * FROM reading_targets"
	rows, err := db.GetDB().Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching reading_targets"))
		return
	}
	defer rows.Close()

	var readingTargets []dto.ReadingTarget
	for rows.Next() {
		var readingTarget dto.ReadingTarget
		err := rows.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages)
		if err != nil {
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
func GetReadingTargetByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingTargetID := vars["id"]

	readingTargetRes, err := getReadingTargetByID(readingTargetID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching reading_target"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readingTargetRes)
}

// getReadingTargetByID retrieves reading_target data from the database by ID.
func getReadingTargetByID(readingTargetID string) (dto.ReadingTarget, error) {
	// Query readingTarget data from the database by ID
	query := "SELECT * FROM reading_target WHERE id = $1"
	row := db.GetDB().QueryRow(query, readingTargetID)

	var readingTarget dto.ReadingTarget
	err := row.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages)
	if err == sql.ErrNoRows {
		return dto.ReadingTarget{}, fmt.Errorf("Reading Target with ID %s not found", readingTargetID)
	} else if err != nil {
		return dto.ReadingTarget{}, err
	}

	return readingTarget, nil
}
