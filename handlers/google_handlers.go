package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthConfig struct {
	GoogleConfig *oauth2.Config
	PostLoginURL string
}

var authConfig AuthConfig

func InitGoogle() {
	authConfig = AuthConfig{
		GoogleConfig: &oauth2.Config{
			ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL: os.Getenv("GOOGLE_CALLBACK_URL"),
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint: google.Endpoint,
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
	code := r.URL.Query().Get("code")
	token, err := authConfig.GoogleConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	_, err = getUserInfo(token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	redirectURL := "" // Change this to your desired success page URL
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
