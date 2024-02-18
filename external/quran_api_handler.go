package external

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/daffashafwan/tadarus-yuk/external/dto"
)

const (
	chaptersEndpoint     = "/chapters/"
	infoEndpoint         = "/info"
	languageQueryParam   = "?language=id"
	versesByPageEndpoint = "/verses/by_page/"
)

var (
	quranAPIURL           string
	quranAPIMaxRetry      int
	quranAPIRetryInterval int
)

func getEnvAsInt(key string) int {
	val, _ := strconv.Atoi(os.Getenv(key))
	return val
}

func InitQuranAPI() {
	quranAPIURL = os.Getenv("QURAN_RAPID_API_URL")
	quranAPIMaxRetry = getEnvAsInt("QURAN_RAPID_MAX_RETRY")
	quranAPIRetryInterval = getEnvAsInt("QURAN_RAPID_RETRY_INTERVAL")
}

func sendAPIRequest(endpoint string) ([]byte, error) {
	url := quranAPIURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}

func retryAPIRequest(endpoint string, result interface{}) error {
	maxTry := quranAPIMaxRetry
	interval := time.Duration(quranAPIRetryInterval) * time.Millisecond

	for attempt := 1; attempt <= maxTry; attempt++ {
		body, err := sendAPIRequest(endpoint)
		if err == nil {
			err = json.Unmarshal(body, result)
			if err == nil {
				return nil
			}
		}

		if attempt == maxTry {
			return errors.New("maximum retries reached")
		}

		time.Sleep(interval)
	}

	return errors.New("maximum retries reached")
}

func GetQuranAPIPages(pageNum string) (dto.QuranAPIPage, error) {
	var quranAPIPage dto.QuranAPIPage
	err := retryAPIRequest(versesByPageEndpoint+pageNum, &quranAPIPage)
	return quranAPIPage, err
}

func GetQuranAPIChapter(chapter string) (dto.QuranAPIChapter, error) {
	var quranAPIChapter dto.QuranAPIChapter
	err := retryAPIRequest(chaptersEndpoint+chapter+languageQueryParam, &quranAPIChapter)
	return quranAPIChapter, err
}

func GetQuranAPIChapterInfo(chapter string) (dto.QuranAPIChapterInfo, error) {
	var quranAPIChapterInfo dto.QuranAPIChapterInfo
	err := retryAPIRequest(chaptersEndpoint+chapter+infoEndpoint+languageQueryParam, &quranAPIChapterInfo)
	return quranAPIChapterInfo, err
}
