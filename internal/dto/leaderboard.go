package dto

import "time"

type Leaderboard struct {
	Type        string    `json:"type"`
	Ranks       []Rank    `json:"ranks"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type Rank struct {
	Username string   `json:"username"`
	Pace     string   `json:"pace"`
	Details  []Detail `json:"details"`
}

type Detail struct {
	ReadingTargetName        string  `json:"readingTargetName"`
	ReadingTargetDescription string  `json:"readingTargetDescription"`
	ReadingTargetDate        string  `json:"readingTargetDate"`
	ReadingTargetProgress    float64 `json:"readingTargetProgress"`
}
