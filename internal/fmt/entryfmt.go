package entryfmt

import (
	"anicliru/internal/api/models"
	"fmt"
	"strconv"
)

func wrapAnimeTitle(anime models.Anime) string {
	if anime.MediaType == "фильм" {
		return fmt.Sprintf("%s (фильм)", anime.Title)
	}

	if anime.EpCtx.TotalEpCount == -1 {
		return fmt.Sprintf("%s (%d из ??? серий)", anime.Title, anime.EpCtx.AiredEpCount)
	}

	if anime.EpCtx.TotalEpCount == anime.EpCtx.AiredEpCount {
		if anime.EpCtx.TotalEpCount == 1 {
			return fmt.Sprintf("%s (%d серия)", anime.Title, anime.EpCtx.TotalEpCount)
		}
		if anime.EpCtx.TotalEpCount < 5 {
			return fmt.Sprintf("%s (%d серии)", anime.Title, anime.EpCtx.TotalEpCount)
		}
		return fmt.Sprintf("%s (%d серий)", anime.Title, anime.EpCtx.TotalEpCount)
	}

	return fmt.Sprintf("%s (%d из %d серий)", anime.Title, anime.EpCtx.AiredEpCount, anime.EpCtx.TotalEpCount)
}

func GetWrappedAnimeTitles(animes []models.Anime) []string {
	wrappedTitles := make([]string, 0, len(animes))
	for _, anime := range animes {
		wrappedTitle := wrapAnimeTitle(anime)
		wrappedTitles = append(wrappedTitles, wrappedTitle)
	}
	return wrappedTitles
}

func EpisodeEntries(epCtx models.EpisodesContext) []string {
	epEntries := make([]string, epCtx.AiredEpCount)
	for i := 0; i < epCtx.AiredEpCount; i++ {
		epEntries[i] = strconv.Itoa(i + 1)
	}

	return epEntries
}
