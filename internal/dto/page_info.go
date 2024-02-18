package dto

type PageInfo struct {
	NumberOfVerses int           `json:"numberOfVerses"`
	AyatSajdah     int           `json:"ayatSajdah"`
	Surahs         []ChapterInfo `json:"surahs"`
}

type ChapterInfo struct {
	Name            string `json:"name"`
	VersesCount     int    `json:"versesCount"`
	RevelationPlace string `json:"revelation_place"`
	Text            string `json:"text"`
	SourceText      string `json:"sourceText"`
}
