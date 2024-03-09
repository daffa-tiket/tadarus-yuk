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
	externalDto "github.com/daffashafwan/tadarus-yuk/external/dto"
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
		err := rows.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages, &readingTarget.Name, &readingTarget.StartPage, &readingTarget.EndPage, &readingTarget.GoogleCalendarID, &readingTarget.IsPublic)
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

	err = updateReadingTarget(readingTarget)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error updating reading target", nil)
		return
	}

	userID, err := authorization.EncryptUserID(readingTarget.UserID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error updating reading target", nil)
		return
	}

	user, err := getUserByID(userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error get user, update reading target", nil)
		return
	}

	event, err := pushCalendarEvent(user.GoogleToken, externalDto.CalendarEvent{
		GoogleCalendarID: readingTarget.GoogleCalendarID,
		EventName:        readingTarget.Name,
		EventDescription: "Membaca Halaman " + strconv.Itoa(readingTarget.StartPage) + " sampai " + strconv.Itoa(readingTarget.EndPage),
		StartDate:        readingTarget.StartDate,
		EndDate:          readingTarget.EndDate,
		Type:             "EDIT",
	})
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error updating reading target calendar", nil)
		return
	}

	readingTarget.GoogleCalendarID = event.Id
	err = updateReadingTarget(readingTarget)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error update, updating reading target calendar google id", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", readingTarget)
}

func updateReadingTarget(readingTarget dto.ReadingTarget) error {
	query := "UPDATE reading_target SET name = $1, start_date = $2, end_date = $3, start_page = $4, end_page = $5, target_pages_per_interval = $6, google_calendar_id = $7, is_public = $8 WHERE target_id = $9"
	_, err := db.GetDB().Exec(query, readingTarget.Name, readingTarget.StartDate, readingTarget.EndDate, readingTarget.StartPage, readingTarget.EndPage, readingTarget.Pages, readingTarget.GoogleCalendarID, readingTarget.IsPublic, readingTarget.ID)
	return err
}

func DeleteReadingTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	readingTargetID := vars["id"]

	readingTarget, err := getReadingTargetByID(readingTargetID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error delete, get reading target", nil)
		return
	}

	userID, err := authorization.EncryptUserID(readingTarget.UserID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error deleting reading target", nil)
		return
	}

	user, err := getUserByID(userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error get user, delet reading target", nil)
		return
	}

	// Delete the user from the database by ID
	query := "DELETE FROM reading_target WHERE target_id = $1"
	_, err = db.GetDB().Exec(query, readingTargetID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error deleting reading target", nil)
		return
	}

	_, err = pushCalendarEvent(user.GoogleToken, externalDto.CalendarEvent{
		GoogleCalendarID: readingTarget.GoogleCalendarID,
		EventName:        "",
		EventDescription: "",
		StartDate:        "2006-01-02",
		EndDate:          "2006-01-02",
		Type:             "DELETE",
	})
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error delete, updating reading target calendar", nil)
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

	user, err := getUserByUsername(userID)
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

	query := "INSERT INTO reading_target (user_id, name, start_date, end_date, start_page, end_page, target_pages_per_interval, google_calendar_id, is_public) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING target_id"
	err = db.GetDB().QueryRow(query, user.ID, readingTarget.Name, readingTarget.StartDate, readingTarget.EndDate, readingTarget.StartPage, readingTarget.EndPage, readingTarget.Pages, readingTarget.GoogleCalendarID, readingTarget.IsPublic).Scan(&readingTarget.ID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error creating reading target", nil)
		return
	}

	event, err := pushCalendarEvent(user.GoogleToken, externalDto.CalendarEvent{
		EventName:        readingTarget.Name,
		EventDescription: "Membaca Halaman " + strconv.Itoa(readingTarget.StartPage) + " sampai " + strconv.Itoa(readingTarget.EndPage),
		StartDate:        readingTarget.StartDate,
		EndDate:          readingTarget.EndDate,
		Type:             "ADD",
	})
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error creating reading target calendar", nil)
		return
	}

	readingTarget.GoogleCalendarID = event.Id
	err = updateReadingTarget(readingTarget)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error updating reading target calendar google id", nil)
		return
	}

	helpers.ResponseJSON(w, err, http.StatusCreated, "SUCCESS", readingTarget)

}

func GetAllReadingTargetByUserID(w http.ResponseWriter, r *http.Request) {
	// Query all reading_targets from the database
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := getUserByUsername(userID)
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
	rows, err := db.GetDB().Query(query, user.ID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error fetching get reading target by userID", nil)
		return
	}
	defer rows.Close()

	var readingTargets []dto.ReadingTarget
	for rows.Next() {
		var readingTarget dto.ReadingTarget
		err := rows.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages, &readingTarget.Name, &readingTarget.StartPage, &readingTarget.EndPage, &readingTarget.GoogleCalendarID, &readingTarget.IsPublic)
		if err != nil {
			helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error scanning reading target rows", nil)
			return
		}
		progresses, err := getReadingProgressByUserIDTargetID(readingTarget.UserID, readingTarget.ID)
		if err != nil {
			helpers.ResponseJSON(w, err, http.StatusInternalServerError, "Error getting reading progress in target", nil)
			return
		}
		readingTarget.LastReadPage = 0
		if len(progresses) > 0 {
			readingTarget.LastReadPage = progresses[0].CurrentPage
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
	err := row.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages, &readingTarget.Name, &readingTarget.StartPage, &readingTarget.EndPage, &readingTarget.GoogleCalendarID, &readingTarget.IsPublic)
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

func getAllPublicReadingTarget(userID int) ([]int,[]dto.ReadingTarget, error) {
	// Query readingTarget data from the database by ID
	var isEligible bool
	query := "SELECT * FROM reading_target WHERE is_public = $1"
	rows, err := db.GetDB().Query(query, true)
	if err != nil {
		return []int{}, []dto.ReadingTarget{}, err
	}
	defer rows.Close()

	var readingTargets []dto.ReadingTarget
	var ids []int
	for rows.Next() {
		var readingTarget dto.ReadingTarget
		err := rows.Scan(&readingTarget.ID, &readingTarget.UserID, &readingTarget.StartDate, &readingTarget.EndDate, &readingTarget.Pages, &readingTarget.Name, &readingTarget.StartPage, &readingTarget.EndPage, &readingTarget.GoogleCalendarID, &readingTarget.IsPublic)
		if err != nil {
			return []int{}, []dto.ReadingTarget{}, err
		}
		ids = append(ids, readingTarget.ID)
		readingTargets = append(readingTargets, readingTarget)
		if readingTarget.UserID == userID {
			isEligible = true
		}
	}

	if !isEligible {
		return []int{}, []dto.ReadingTarget{}, errors.New("user didn't have any public reading target")
	}

	return ids, readingTargets, nil
}