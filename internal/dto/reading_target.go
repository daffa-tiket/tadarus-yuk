package dto

type ReadingTarget struct {
	ID        int `json:"id"`
	UserID    int `json:"userId"`
	StartDate int `json:"startDate"`
	EndDate   int `json:"endDate"`
	Pages     int `json:"Pages"`
}
