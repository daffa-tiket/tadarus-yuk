package dto

type ReadingTarget struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	UserID           int     `json:"userId"`
	StartDate        string  `json:"startDate"`
	EndDate          string  `json:"endDate"`
	StartPage        int     `json:"startPage"`
	EndPage          int     `json:"endPage"`
	Pages            float64 `json:"pages"`
	Progress         float64 `json:"progress"`
	LastReadPage     int     `json:"lastReadPage"`
	GoogleCalendarID string  `json:"-"`
	IsPublic         bool    `json:"isPublic"`
}

type ReadingTargetWithUser struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	UserID           int     `json:"userId"`
	StartDate        string  `json:"startDate"`
	EndDate          string  `json:"endDate"`
	StartPage        int     `json:"startPage"`
	EndPage          int     `json:"endPage"`
	Pages            float64 `json:"pages"`
	Progress         float64 `json:"progress"`
	LastReadPage     int     `json:"lastReadPage"`
	GoogleCalendarID string  `json:"-"`
	IsPublic         bool    `json:"isPublic"`
	User             User
}
