package animefmt

import (
	"anicliru/internal/api/models"
	"fmt"
	"sort"
	"strconv"
)

func wrapAnimeTitle(anime models.Anime) string {
	if anime.TotalEpCount == len(anime.Episodes) {
		return fmt.Sprintf("%s (%d серий)", anime.Title, anime.TotalEpCount)
	}
	return fmt.Sprintf("%s (%d из %d серий)", anime.Title, len(anime.Episodes), anime.TotalEpCount)
}

func GetWrappedAnimeTitles(animes []models.Anime) []string {
	wrappedTitles := make([]string, 0, len(animes))
	for _, anime := range animes {
		wrappedTitle := wrapAnimeTitle(anime)
		wrappedTitles = append(wrappedTitles, wrappedTitle)
	}
	return wrappedTitles
}

func GetEpisodes(anime *models.Anime) []string{
	var episodes []string

	keys := make([]int, 0, len(anime.Episodes))
	for k := range anime.Episodes {
		keys = append(keys, k)
	}
	sort.Sort(sort.IntSlice(keys))

	for _, k := range keys {
		episodes = append(episodes, strconv.Itoa(k))
	}

    return episodes
}
