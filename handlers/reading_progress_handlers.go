package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/daffashafwan/tadarus-yuk/db"
	"github.com/daffashafwan/tadarus-yuk/internal/dto"
	"github.com/daffashafwan/tadarus-yuk/internal/helpers"
	"github.com/gorilla/mux"
)

func GetAllReadingProgress(w http.ResponseWriter, r *http.Request) {
	// Query all reading_progress from the database
	query := "SELECT * FROM reading_progress"
	rows, err := db.GetDB().Query(query)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching reading progress", nil)
		return
	}
	defer rows.Close()

	var readingProgresss []dto.ReadingProgress
	for rows.Next() {
		var readingProgress dto.ReadingProgress
		err := rows.Scan(&readingProgress.ID, &readingProgress.UserID, &readingProgress.TargetID, &readingProgress.CurrentPage, &readingProgress.TimeStamp)
		if err != nil {
			helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching get all reading progress", nil)
			return
		}
		readingProgresss = append(readingProgresss, readingProgress)
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", readingProgresss)
	
}

// GetReadingProgressByIDHandler handles requests to get a reading_progress by ID.
func GetReadingProgressByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingProgressID := vars["id"]

	readingProgressRes, err := getReadingProgressByID(readingProgressID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching reading progress by ID", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", readingProgressRes)
}

func UpdateReadingProgressByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingProgressID := vars["id"]

	readingProgress, err := getReadingProgressByID(readingProgressID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching reading progress by ID", nil)
		return
	}

	var readingProgressUpdate dto.ReadingProgress
	err = json.NewDecoder(r.Body).Decode(&readingProgressUpdate)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	readingProgress.CurrentPage = readingProgressUpdate.CurrentPage

	query := "UPDATE reading_progress SET current_page = $1 WHERE progress_id = $2"
	_, err = db.GetDB().Exec(query, readingProgress.CurrentPage, readingProgress.ID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error updating reading progress", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", readingProgress)
}

func DeleteReadingProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingProgressID := vars["id"]

	// Delete the user from the database by ID
	query := "DELETE FROM reading_progress WHERE progress_id = $1"
	_, err := db.GetDB().Exec(query, readingProgressID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error deleting reading progress", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusNoContent, "SUCCESS", nil)
}

func CreateReadingProgress(w http.ResponseWriter, r *http.Request) {
	var readingProgress dto.ReadingProgress
	err := json.NewDecoder(r.Body).Decode(&readingProgress)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]
	targetID := vars["tid"]

	user, err := getUserByUsername(userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Error get user", nil)
		return
	}

	readingTarget, err := getReadingTargetByID(targetID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Error get reading target", nil)
		return
	}

	if readingProgress.CurrentPage < readingTarget.StartPage || readingProgress.CurrentPage > readingTarget.EndPage {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "page is more or less than target", nil)
		return
	}

	readedPage := getReadedPages(user.ID, targetID)
	if containsValue(readedPage, readingProgress.CurrentPage) {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Page "+ strconv.Itoa(readingProgress.CurrentPage) +"  already read", nil)
		return
	}

	readingProgress.UserID = user.ID
	readingProgress.TargetID = readingTarget.ID

	query := "INSERT INTO reading_progress (user_id, target_id, current_page) VALUES ($1, $2, $3) RETURNING progress_id"
	err = db.GetDB().QueryRow(query, readingProgress.UserID, readingProgress.TargetID, readingProgress.CurrentPage).Scan(&readingProgress.ID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error creating reading progress", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusCreated, "SUCCESS", readingProgress)
	
}

func GetAllReadingProgressByUserID(w http.ResponseWriter, r *http.Request) {
	// Query all reading_targets from the database
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := getUserByUsername(userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error decrypting", nil)
		return
	}

	query := "SELECT * FROM reading_progress where user_id = $1 ORDER BY last_update_timestamp ASC"
	rows, err := db.GetDB().Query(query, user.ID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching get all reading progress", nil)
		return
	}
	defer rows.Close()

	var readingProgresses []dto.ReadingProgress
	readingProgressSorted := make(map[int]map[string][]dto.ReadingProgress)
	for rows.Next() {
		var readingProgress dto.ReadingProgress
		err := rows.Scan(&readingProgress.ID, &readingProgress.UserID, &readingProgress.TargetID, &readingProgress.CurrentPage, &readingProgress.TimeStamp)
		if err != nil {
			helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error scanning all reading progress by userID", nil)
			return
		}
		if readingProgressSorted[readingProgress.TimeStamp.Year()] == nil {
			readingProgressSorted[readingProgress.TimeStamp.Year()] = make(map[string][]dto.ReadingProgress)
		}
		if readingProgressSorted[readingProgress.TimeStamp.Year()][readingProgress.TimeStamp.Month().String()] == nil {
			readingProgressSorted[readingProgress.TimeStamp.Year()][readingProgress.TimeStamp.Month().String()] = []dto.ReadingProgress{}
		}
		
		readingProgressSorted[readingProgress.TimeStamp.Year()][readingProgress.TimeStamp.Month().String()] = append(
			readingProgressSorted[readingProgress.TimeStamp.Year()][readingProgress.TimeStamp.Month().String()],
			readingProgress,
		)
		readingProgresses = append(readingProgresses, readingProgress)
	}

	response := dto.ReadingProgressAggregated{
		ReadingProgress: readingProgresses,
		ReadingProgressSorted: readingProgressSorted,
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", response)
}

func GetAllReadingProgressByUserIDTargetID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	targetID := vars["tid"]

	user, err := getUserByUsername(userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Error get user", nil)
		return
	}

	readingTarget, err := getReadingTargetByID(targetID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusBadRequest, "Error get reading target", nil)
		return
	}

	readingProgress, err := getReadingProgressByUserIDTargetID(user.ID, readingTarget.ID)

	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error get reading progress", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", readingProgress)
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
		log.Printf("Error : %v", err.Error())
		return dto.ReadingProgress{}, err
	}

	return readingProgress, nil
}

func getReadedPages(userID int, targetID string) []int {
	readedPages := make([]int, 0)

	targetIDConv, _ := strconv.Atoi(targetID)
	readingProgress, err := getReadingProgressByUserIDTargetID(userID, targetIDConv)

	if err != nil {
		log.Printf("Error : %v", err.Error())
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

	query := "SELECT * FROM reading_progress WHERE user_id = $1 AND target_id = $2 ORDER BY last_update_timestamp DESC"
	rows, err := db.GetDB().Query(query, userID, targetID)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		return []dto.ReadingProgress{}, err
	}
	defer rows.Close()

	var readingProgresss []dto.ReadingProgress
	for rows.Next() {
		var readingProgress dto.ReadingProgress
		err := rows.Scan(&readingProgress.ID, &readingProgress.UserID, &readingProgress.TargetID, &readingProgress.CurrentPage, &readingProgress.TimeStamp)
		if err != nil {
			log.Printf("Error : %v", err.Error())
			return []dto.ReadingProgress{}, err
		}
		readingProgresss = append(readingProgresss, readingProgress)
	}

	return readingProgresss, nil
}

func getReadingProgressByTargetIDsAndTimeRange(targetIDs []int, startTime, endTime time.Time) ([]dto.ReadingProgress, error) {
    inClause := helpers.BuildInClause(targetIDs)

	if len(inClause) == 0 {
		return []dto.ReadingProgress{}, errors.New("empty list of target IDs")
	}
	query := fmt.Sprintf(`
        SELECT * FROM reading_progress
        WHERE target_id IN (%s)
        AND last_update_timestamp >= $1 AND last_update_timestamp <= $2
    `, inClause)

	fmt.Println(startTime.String())

    rows, err := db.GetDB().Query(query, startTime, endTime)
    if err != nil {
        log.Printf("Error: %v", err.Error())
        return []dto.ReadingProgress{}, err
    }
    defer rows.Close()

    var readingProgresss []dto.ReadingProgress
    for rows.Next() {
        var readingProgress dto.ReadingProgress
        err := rows.Scan(&readingProgress.ID, &readingProgress.UserID, &readingProgress.TargetID, &readingProgress.CurrentPage, &readingProgress.TimeStamp)
        if err != nil {
            log.Printf("Error: %v", err.Error())
            return []dto.ReadingProgress{}, err
        }
        readingProgresss = append(readingProgresss, readingProgress)
    }

    return readingProgresss, nil
}
