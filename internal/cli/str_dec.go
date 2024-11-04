package cli

import (
	"anicliru/internal/api"
	"fmt"
	"strings"
)

func decoratedAnimeTitles(animeInfoList []api.AnimeInfo) []string {
	foundAnimeTitles := make([]string, 0, len(animeInfoList))

	for i, animeInfo := range animeInfoList {
		episodeCount := len(animeInfo.Media.Episodes)
		var b strings.Builder

		fmt.Fprintf(&b, "%d %s", i+1, animeInfo.Names.RU)
		if episodeCount == 1 {
			fmt.Fprintf(&b, " (%d %s)", episodeCount, "эпизод")
		} else if episodeCount > 1 {
			fmt.Fprintf(&b, " (%d %s)", episodeCount, "эпизодов")
		} else {
			fmt.Fprintf(&b, " (Нет доступных серий)")
		}
		foundAnimeTitles = append(foundAnimeTitles, b.String())
	}
	return foundAnimeTitles
}
