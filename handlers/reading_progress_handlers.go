package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daffashafwan/tadarus-yuk/db"
	"github.com/daffashafwan/tadarus-yuk/internal/dto"
	"github.com/gorilla/mux"
)

func GetAllReadingProgress(w http.ResponseWriter, r *http.Request) {
	// Query all reading_progress from the database
	query := "SELECT * FROM reading_progress"
	rows, err := db.GetDB().Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching reading_progress"))
		return
	}
	defer rows.Close()

	var readingProgresss []dto.ReadingProgress
	for rows.Next() {
		var readingProgress dto.ReadingProgress
		err := rows.Scan(&readingProgress.ID, &readingProgress.UserID, &readingProgress.TargetID, &readingProgress.CurrentPage, &readingProgress.TimeStamp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error scanning reading_progress rows"))
			return
		}
		readingProgresss = append(readingProgresss, readingProgress)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readingProgresss)
}

// GetReadingProgressByIDHandler handles requests to get a reading_progress by ID.
func GetReadingProgressByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingProgressID := vars["id"]

	readingProgressRes, err := getReadingProgressByID(readingProgressID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching reading_progress"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readingProgressRes)
}

func UpdateReadingProgressByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingProgressID := vars["id"]

	readingProgress, err := getReadingProgressByID(readingProgressID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching reading progress"))
		return
	}

	var readingProgressUpdate dto.ReadingProgress
	err = json.NewDecoder(r.Body).Decode(&readingProgressUpdate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}

	readingProgress.CurrentPage = readingProgressUpdate.CurrentPage

	query := "UPDATE reading_progress SET current_page = $1 WHERE progress_id = $2"
	_, err = db.GetDB().Exec(query, readingProgress.CurrentPage, readingProgress.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error updating readingProgress data"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readingProgress)
}

func DeleteReadingProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingProgressID := vars["id"]

	// Delete the user from the database by ID
	query := "DELETE FROM reading_progress WHERE progress_id = $1"
	_, err := db.GetDB().Exec(query, readingProgressID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error deleting reading progress"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func CreateReadingProgress(w http.ResponseWriter, r *http.Request) {
	var readingProgress dto.ReadingProgress
	err := json.NewDecoder(r.Body).Decode(&readingProgress)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]
	targetID := vars["tid"]

	user, err := getUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error get user"))
		return
	}

	readingTarget, err := getReadingTargetByID(targetID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error get reading target"))
		return
	}

	if readingProgress.CurrentPage < readingTarget.StartPage || readingProgress.CurrentPage > readingTarget.EndPage {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("page is more or less than target"))
		return
	}

	readedPage := getReadedPages(userID, targetID)
	if containsValue(readedPage, readingProgress.CurrentPage){
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Page already read"))
		return 
	}

	readingProgress.UserID = user.ID
	readingProgress.TargetID = readingTarget.ID

	query := "INSERT INTO reading_progress (user_id, target_id, current_page) VALUES ($1, $2, $3) RETURNING progress_id"
	err = db.GetDB().QueryRow(query, readingProgress.UserID, readingProgress.TargetID, readingProgress.CurrentPage).Scan(&readingProgress.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error creating reading progress"))
		return
	}

	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(readingProgress)
}

func GetAllReadingProgressByUserID(w http.ResponseWriter, r *http.Request) {
	// Query all reading_targets from the database
	vars := mux.Vars(r)
	userID := vars["id"]

	query := "SELECT * FROM reading_progress where user_id = $1"
	rows, err := db.GetDB().Query(query, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching reading progress"))
		return
	}
	defer rows.Close()

	var readingProgresses []dto.ReadingProgress
	for rows.Next() {
		var readingProgress dto.ReadingProgress
		err := rows.Scan(&readingProgress.ID, &readingProgress.UserID, &readingProgress.TargetID, &readingProgress.CurrentPage, &readingProgress.TimeStamp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error scanning reading progress rows"))
			return
		}
		readingProgresses = append(readingProgresses, readingProgress)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readingProgresses)
}

func GetAllReadingProgressByUserIDTargetID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	targetID := vars["tid"]

	user, err := getUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error get user"))
		return
	}

	readingTarget, err := getReadingTargetByID(targetID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error get reading target"))
		return
	}

	readingProgress, err := getReadingProgressByUserIDTargetID(user.ID, readingTarget.ID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error get reading progress"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(readingProgress)
}

// getReadingProgressByID retrieves reading_progress data from the database by ID.
func getReadingProgressByID(readingProgressID string) (dto.ReadingProgress, error) {
	// Query readingProgress data from the database by ID
	query := "SELECT * FROM reading_progress WHERE progress_id = $1"
	row := db.GetDB().QueryRow(query, readingProgressID)

	var readingProgress dto.ReadingProgress
	err := row.Scan(&readingProgress.ID, &readingProgress.UserID, &readingProgress.TargetID, &readingProgress.CurrentPage, &readingProgress.TimeStamp)
	if err == sql.ErrNoRows {
		return dto.ReadingProgress{}, fmt.Errorf("Reading Target with ID %s not found", readingProgressID)
	} else if err != nil {
		return dto.ReadingProgress{}, err
	}

	return readingProgress, nil
}

func getReadedPages(userID, targetID string) []int {
	readedPages := make([]int, 0)

	userIDConv, _ := strconv.Atoi(userID)
	targetIDConv, _ := strconv.Atoi(targetID)
	readingProgress, err := getReadingProgressByUserIDTargetID(userIDConv, targetIDConv)

	if err != nil {
		return readedPages
	}

	for _, v := range readingProgress {
		readedPages = append(readedPages, v.CurrentPage)
	}

	return readedPages
}

func containsValue(slice []int, value int) bool {
	for _, element := range slice {
		if element == value {
			return true
		}
	}
	return false
}

func getReadingProgressByUserIDTargetID(userID, targetID int) ([]dto.ReadingProgress, error) {

	query := "SELECT * FROM reading_progress WHERE user_id = $1 AND target_id = $2"
	rows, err := db.GetDB().Query(query, userID, targetID)
	if err != nil {
		return []dto.ReadingProgress{}, err
	}
	defer rows.Close()

	var readingProgresss []dto.ReadingProgress
	for rows.Next() {
		var readingProgress dto.ReadingProgress
		err := rows.Scan(&readingProgress.ID, &readingProgress.UserID, &readingProgress.TargetID, &readingProgress.CurrentPage, &readingProgress.TimeStamp)
		if err != nil {
			return []dto.ReadingProgress{}, err
		}
		readingProgresss = append(readingProgresss, readingProgress)
	}

	return readingProgresss, nil
}