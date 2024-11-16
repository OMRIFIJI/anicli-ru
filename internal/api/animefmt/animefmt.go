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
