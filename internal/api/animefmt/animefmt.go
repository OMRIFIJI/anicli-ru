package animefmt

import (
    "anicliru/internal/api/types"
    "fmt"
)


func WrapAnimeTitle(anime types.Anime, i int) string {
    wrappedTitle := fmt.Sprintf("%d ", i+1)
    if anime.IsAvailable {
        if anime.TotalEpCount == len(anime.Episodes) {
            wrappedTitle += fmt.Sprintf("%s (%d серий)", anime.Title, anime.TotalEpCount)
        } else {
            wrappedTitle += fmt.Sprintf("%s (%d из %d серий)", anime.Title, len(anime.Episodes), anime.TotalEpCount)
        }
        return wrappedTitle
    }

    if anime.IsRegionBlock {
        wrappedTitle += anime.Title + " (не доступно в вашем регионе)"
    } else {
        wrappedTitle += anime.Title + " (не удалось распарсить)"
    }

    return wrappedTitle
}
