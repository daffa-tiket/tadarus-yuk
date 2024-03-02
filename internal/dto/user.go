package dto

type User struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	GoogleToken string `json:"-"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Admin struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
