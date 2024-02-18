package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/daffashafwan/tadarus-yuk/db"
	"github.com/daffashafwan/tadarus-yuk/internal/authorization"
	"github.com/daffashafwan/tadarus-yuk/internal/dto"
	"github.com/daffashafwan/tadarus-yuk/internal/helpers"
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
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching get all reading target", nil)
		return
	}
	defer rows.Close()

	var readingTargets []dto.ReadingTarget
	for rows.Next() {
		var readingTarget dto.ReadingTarget
		err := rows.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages, &readingTarget.Name, &readingTarget.StartPage, &readingTarget.EndPage)
		if err != nil {
			helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching reading target rows", nil)
			return
		}
		readingTargets = append(readingTargets, readingTarget)
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", readingTargets)
}

// GetReadingTargetByIDHandler handles requests to get a reading_target by ID.
func GetReadingTargetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingTargetID := vars["id"]

	readingTargetRes, err := getReadingTargetByID(readingTargetID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching get reading target by ID", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", readingTargetRes)
}

func UpdateReadingTargetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingTargetID := vars["id"]

	readingTarget, err := getReadingTargetByID(readingTargetID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching get reading target by ID", nil)
		return
	}

	var readingTargetUpdate dto.ReadingTarget
	err = json.NewDecoder(r.Body).Decode(&readingTargetUpdate)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	readingTarget.Name = readingTargetUpdate.Name
	readingTarget.StartDate = readingTargetUpdate.StartDate
	readingTarget.EndDate = readingTargetUpdate.EndDate
	readingTarget.Pages = readingTargetUpdate.Pages

	query := "UPDATE reading_target SET name = $1, start_date = $2, end_date = $3, start_page = $4, end_page = $5, target_pages_per_interval = $6 WHERE target_id = $7"
	_, err = db.GetDB().Exec(query, readingTarget.Name, readingTarget.StartDate, readingTarget.EndDate, readingTarget.StartPage, readingTarget.EndPage, readingTarget.Pages, readingTarget.ID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error updating reading target", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", readingTarget)
}

func DeleteReadingTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingTargetID := vars["id"]

	// Delete the user from the database by ID
	query := "DELETE FROM reading_target WHERE target_id = $1"
	_, err := db.GetDB().Exec(query, readingTargetID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error deleting reading target", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusNoContent, "SUCCESS", nil)
}

func CreateReadingTargetByUserID(w http.ResponseWriter, r *http.Request) {
	var readingTarget dto.ReadingTarget
	err := json.NewDecoder(r.Body).Decode(&readingTarget)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := getUserByID(userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Error user not found", nil)
		return
	}

	if !isValidDateRange(readingTarget.StartDate, readingTarget.EndDate) {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "invalid date or date range", nil)
		return
	}

	if readingTarget.Pages < 1 || readingTarget.Pages > PagesAlQuran {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "invalid number of pages", nil)
		return
	}

	query := "INSERT INTO reading_target (user_id, name, start_date, end_date, start_page, end_page, target_pages_per_interval) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING target_id"
	err = db.GetDB().QueryRow(query, user.ID, readingTarget.Name, readingTarget.StartDate, readingTarget.EndDate, readingTarget.StartPage, readingTarget.EndPage, readingTarget.Pages).Scan(&readingTarget.ID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error creating reading target", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusCreated, "SUCCESS", readingTarget)

}

func GetAllReadingTargetByUserID(w http.ResponseWriter, r *http.Request) {
	// Query all reading_targets from the database
	vars := mux.Vars(r)
	userID := vars["id"]

	decrypted, err := authorization.DecryptUserID(userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error decrypting process", nil)
		return
	}

	//query := "SELECT * FROM reading_target where user_id = $1"
	query := `
        SELECT *
        FROM reading_target
        WHERE user_id = $1;
    `
	rows, err := db.GetDB().Query(query, decrypted)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching get reading target by userID", nil)
		return
	}
	defer rows.Close()

	var readingTargets []dto.ReadingTarget
	for rows.Next() {
		var readingTarget dto.ReadingTarget
		err := rows.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages, &readingTarget.Name, &readingTarget.StartPage, &readingTarget.EndPage)
		if err != nil {
			helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error scanning reading target rows", nil)
			return
		}
		progresses, err := getReadingProgressByUserIDTargetID(readingTarget.UserID, readingTarget.ID)
		if err != nil {
			helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error getting reading progress in target", nil)
			return
		}
		readingTarget.Progress = float64(len(progresses)) / readingTarget.Pages * 100
		readingTarget.Progress = float64(int(readingTarget.Progress*10)) / 10
		readingTargets = append(readingTargets, readingTarget)
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", readingTargets)
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
