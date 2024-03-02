package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	externalDto "github.com/daffashafwan/tadarus-yuk/external/dto"
	"github.com/daffashafwan/tadarus-yuk/internal/authorization"
	"github.com/daffashafwan/tadarus-yuk/internal/dto"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type AuthConfig struct {
	GoogleConfig *oauth2.Config
	PostLoginURL string
}

var authConfig AuthConfig

func InitGoogle() {
	authConfig = AuthConfig{
		GoogleConfig: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
			Scopes:       []string{"openid", "profile", "email", "https://www.googleapis.com/auth/calendar"},
			Endpoint:     google.Endpoint,
		},
		PostLoginURL: os.Getenv("POST_LOGIN_URL"),
	}
}

type userInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := authConfig.GoogleConfig.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	var isFirstLogin = "true"
	code := r.URL.Query().Get("code")
	token, err := authConfig.GoogleConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	user, err := getUserInfo(token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	userByEmail, _ := getUserByEmail(user.Email)
	if userByEmail.Email == "" {
		err = createUser(dto.User{
			Username:    user.ID,
			Email:       user.Email,
			GoogleToken: token.AccessToken,
		}, "")
		if err != nil {
			http.Error(w, "Failed create user", http.StatusInternalServerError)
			return
		}
	} else {
		userByEmail.GoogleToken = token.AccessToken

		err = updateUser(userByEmail)
		if err != nil {
			http.Error(w, "Failed set user token", http.StatusInternalServerError)
			return
		}

		isFirstLogin = "false"
	}

	authToken, err := authorization.GenerateAuthToken(userByEmail.ID, "user")
	if err != nil {
		http.Error(w, "Failed to auth", http.StatusInternalServerError)
		return
	}

	redirectURL := authConfig.PostLoginURL + "?user=" + url.QueryEscape(userByEmail.Username) + "&token=" + authToken + "&isFirstLogin=" + isFirstLogin // Change this to your desired success page URL
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func getUserInfo(accessToken string) (*userInfo, error) {
	resp, err := http.Get(fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?access_token=%s", accessToken))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user userInfo
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func getClientWithAccessToken(accessToken string) *http.Client {
	token := &oauth2.Token{AccessToken: accessToken}
	config := &oauth2.Config{}

	return config.Client(context.Background(), token)
}

func pushCalendarEvent(accessToken string, calendarEvent externalDto.CalendarEvent) (*calendar.Event, error) {
	// Create a new Calendar service
	var eventCreated *calendar.Event
	googleClient := getClientWithAccessToken(accessToken)
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(googleClient))
	if err != nil {
		return nil, err
	}

	jakartaLocation, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}

	startDate, err := time.Parse("2006-01-02", calendarEvent.StartDate)
	if err != nil {
		return nil, err
	}
	desiredTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 21, 0, 0, 0, jakartaLocation)
	startDateFormat := desiredTime.Format(time.RFC3339)

	endDate, err := time.Parse("2006-01-02", calendarEvent.EndDate)
	if err != nil {
		return nil, err
	}
	endDateFormat := desiredTime.Add(1 * time.Hour).Format(time.RFC3339)

	daysLength := int(endDate.Sub(startDate).Hours() / 24) + 1

	event := &calendar.Event{
		Summary:     calendarEvent.EventName + " [Update Progress di App Tadaroosh]",
		Description: calendarEvent.EventDescription,
		Start: &calendar.EventDateTime{
			DateTime: startDateFormat,
			TimeZone: "Asia/Jakarta",
		},
		End: &calendar.EventDateTime{
			DateTime: endDateFormat,
			TimeZone: "Asia/Jakarta",
		},
		Recurrence: []string{"RRULE:FREQ=DAILY;COUNT=" + strconv.Itoa(daysLength)},
	}

	event.Reminders = &calendar.EventReminders{
		UseDefault: true,
	}

	switch calendarEvent.Type {
	case "ADD":
		createdEvent, err := srv.Events.Insert("primary", event).Do()
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		eventCreated = createdEvent
	case "EDIT":
		updatedEvent, err := srv.Events.Update("primary", calendarEvent.GoogleCalendarID, event).Do()
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		eventCreated = updatedEvent
	case "DELETE":
		err = srv.Events.Delete("primary", calendarEvent.GoogleCalendarID).Do()
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
	default:
		return nil, errors.New("Error push calendar")
	}

	return eventCreated, nil
}
