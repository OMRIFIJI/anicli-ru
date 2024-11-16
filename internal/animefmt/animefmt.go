package animefmt

import (
	"anicliru/internal/api/types"
	"fmt"
)

func WrapAnimeTitle(anime types.Anime, i int) string {
	if anime.TotalEpCount == len(anime.Episodes) {
        return fmt.Sprintf("%s (%d серий)", anime.Title, anime.TotalEpCount)
	}
    return fmt.Sprintf("%s (%d из %d серий)", anime.Title, len(anime.Episodes), anime.TotalEpCount)
}


func GetWrappedAnimeTitles(animes []types.Anime) []string {
	wrappedTitles := make([]string, 0, len(animes))
	for i, anime := range animes {
		wrappedTitle := WrapAnimeTitle(anime, i)
		wrappedTitles = append(wrappedTitles, wrappedTitle)
	}
	return wrappedTitles
}
