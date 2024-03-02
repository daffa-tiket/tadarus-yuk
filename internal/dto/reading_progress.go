package dto

import "time"

type ReadingProgress struct {
	ID          int       `json:"id"`
	UserID      int       `json:"userId"`
	TargetID    int       `json:"targetID"`
	CurrentPage int       `json:"currentPage"`
	TimeStamp   time.Time `json:"timeStamp"`
}

type ReadingProgressAggregated struct {
	ReadingProgress []ReadingProgress `json:"readingProgress"`
	ReadingProgressSorted map[int]map[string][]ReadingProgress `json:"readingProgressSorted"`
}
