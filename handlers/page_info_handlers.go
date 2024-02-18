package handlers

import (
	"log"
	"strconv"
	"strings"

	"github.com/daffashafwan/tadarus-yuk/external"
	"github.com/daffashafwan/tadarus-yuk/internal/dto"
)

func getPageInfo(page string)(dto.PageInfo, error){
	var pageInfo dto.PageInfo
	surahMap := make(map[string]dto.ChapterInfo)
	pageRes, err := external.GetQuranAPIPages(page)
	if err != nil {
		log.Printf("[GetQuranCloudPages] error get page %s, with error : %s", page, err.Error())
		return dto.PageInfo{}, err
	}
	pageInfo.NumberOfVerses = len(pageRes.Verses)
	currentChapter := "0"
	for _, verse  := range pageRes.Verses {
		result := strings.Split(verse.VerseKey, ":")
		currentChapter = result[0]
		surah, err := external.GetQuranAPIChapter(currentChapter)
		if err != nil {
			log.Printf("[GetQuranAPIChapter] error get chapter %s from page %s, with error : %s", currentChapter, page, err.Error())
			continue
		}
		chapterInfo := dto.ChapterInfo{
			Name: surah.Chapter.NameSimple,
			VersesCount: surah.Chapter.VersesCount,
			RevelationPlace: surah.Chapter.RevelationPlace,
		}

		surahInfo, err := external.GetQuranAPIChapterInfo(strconv.Itoa(surah.Chapter.ID))
		if err != nil {
			log.Printf("[GetQuranAPIChapter] error get info from chapter %s from page %s, with error : %s", currentChapter, page, err.Error())
			continue
		}
		chapterInfo.SourceText = surahInfo.ChapterInfo.Source
		chapterInfo.Text = surahInfo.ChapterInfo.Text
		if surahInfo.ChapterInfo.ShortText != "" {
			chapterInfo.Text = surahInfo.ChapterInfo.ShortText
		}
		surahMap[chapterInfo.Name] = chapterInfo	
	}
	pageInfo.Surahs = surahMap

	return pageInfo, nil
}