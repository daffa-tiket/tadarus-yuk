package dto
type QuranAPIChapterInfo struct {
	ChapterInfo ChapterInfo `json:"chapter_info"`
}

type ChapterInfo struct {
	ID           int    `json:"id"`
	ChapterID    int    `json:"chapter_id"`
	LanguageName string `json:"language_name"`
	ShortText    string `json:"short_text"`
	Source       string `json:"source"`
	Text         string `json:"text"`
}

type Verse struct {
	ID               int     `json:"id"`
	VerseNumber      int     `json:"verse_number"`
	VerseKey         string  `json:"verse_key"`
	HizbNumber       int     `json:"hizb_number"`
	RubElHizbNumber  int     `json:"rub_el_hizb_number"`
	RukuNumber       int     `json:"ruku_number"`
	ManzilNumber     int     `json:"manzil_number"`
	SajdahNumber     *int    `json:"sajdah_number"`
	PageNumber       int     `json:"page_number"`
	JuzNumber        int     `json:"juz_number"`
}

type Pagination struct {
	PerPage      int `json:"per_page"`
	CurrentPage  int `json:"current_page"`
	NextPage     *int `json:"next_page"`
	TotalPages   int `json:"total_pages"`
	TotalRecords int `json:"total_records"`
}

type QuranAPIPage struct {
	Verses     []Verse     `json:"verses"`
	Pagination Pagination `json:"pagination"`
}

type QuranAPIChapter struct {
	Chapter Chapter `json:"chapter"`
}

type Chapter struct {
	ID              int    `json:"id"`
	RevelationPlace string `json:"revelation_place"`
	RevelationOrder int    `json:"revelation_order"`
	BismillahPre    bool   `json:"bismillah_pre"`
	NameSimple      string `json:"name_simple"`
	NameComplex     string `json:"name_complex"`
	NameArabic      string `json:"name_arabic"`
	VersesCount     int    `json:"verses_count"`
	Pages           []int  `json:"pages"`
	TranslatedName  struct {
		LanguageName string `json:"language_name"`
		Name         string `json:"name"`
	} `json:"translated_name"`
}