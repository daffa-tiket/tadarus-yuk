package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/daffashafwan/tadarus-yuk/internal/dto"
	"github.com/daffashafwan/tadarus-yuk/internal/helpers"
)

var (
	leaderboardCache = make(map[string]dto.Leaderboard)
)

func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	leaderboardType := queryParams.Get("type")
	userID := queryParams.Get("userID")

	progress := make(map[int]int)
	now := time.Now()

	var startTime, endTime time.Time
	var divider float64
	switch leaderboardType {
	case "daily":
		endTime = now
		startTime = endTime.AddDate(0, 0, -1)
		divider = 1
	case "weekly":
		endTime = now
		startTime = endTime.AddDate(0, 0, -6)
		divider = 7
	case "last30days":
		endTime = now
		startTime = endTime.AddDate(0, 0, -30)
		divider = 30
	default:
		helpers.ResponseJSON(w, errors.New("leaderboard type is not valid"), http.StatusBadRequest, "Error get leaderboard", nil)
		return
	}

	user, err := getUserByUsername(userID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "[leaderboard] Error get user", nil)
		return
	}

	ids, readingTargets, err := getAllPublicReadingTarget(user.ID)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "[leaderboard] Error get leaderboard", nil)
		return
	}
	readingProgress, err := getReadingProgressByTargetIDsAndTimeRange(ids, startTime, endTime)
	if err != nil {
		helpers.ResponseJSON(w, err, http.StatusInternalServerError, "[leaderboard] Error get leaderboard", nil)
		return
	}

	for _, rp := range readingProgress {
		progress[rp.UserID]++
	}

	var progressSlice []struct {
		Key   int
		Value int
	}

	for k, v := range progress {
		progressSlice = append(progressSlice, struct {
			Key   int
			Value int
		}{k, v})
	}

	sort.Slice(progressSlice, func(i, j int) bool {
		return progressSlice[i].Value > progressSlice[j].Value
	})

	ranks := make([]dto.Rank, 0)
	for _, val := range progressSlice {
		details := getReadingTargetByUserIDForLeaderboard(val.Key, readingTargets)
		user, _ := getUserByIDWithoutEncrypt(val.Key)
		pace := float64(val.Value) / divider
		paceFormatted := fmt.Sprintf("%.3f", pace)
		ranks = append(ranks, dto.Rank{
			Username: user.DisplayName,
			Pace:     paceFormatted,
			Details:  details,
		})
	}

	if _, ok := leaderboardCache[leaderboardType]; !ok {
		leaderboardCache[leaderboardType] = dto.Leaderboard{}
	}

	leaderboardCache[leaderboardType] = dto.Leaderboard{
		Type:        leaderboardType,
		Ranks:       ranks,
		LastUpdated: now,
	}

	helpers.ResponseJSON(w, err, http.StatusOK, "SUCCESS", leaderboardCache[leaderboardType])
}

func getReadingTargetByUserIDForLeaderboard(userID int, readingTarget []dto.ReadingTarget) []dto.Detail {
	var result []dto.Detail
	for _, rt := range readingTarget {
		var isEligible bool
		progress, _ := getReadingProgressByUserIDTargetID(userID, rt.ID)
		if len(progress) > 0 {
			isEligible = true
		}
		if rt.UserID == userID && isEligible{
			startDate := strings.Split(rt.StartDate, "T")
			endDate := strings.Split(rt.EndDate, "T")
			result = append(result, dto.Detail{
				ReadingTargetName:        rt.Name,
				ReadingTargetDescription: "Halaman " + strconv.Itoa(rt.StartPage) + " - " + strconv.Itoa(rt.EndPage),
				ReadingTargetDate:        "Mulai : " + startDate[0] + ", Selesai : " + endDate[0],
				ReadingTargetProgress:    rt.Progress,
			})
		}
	}

	return result
}
