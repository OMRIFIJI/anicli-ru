package cli

import (
	"anicliru/internal/api"
	"fmt"
	"strconv"
	"strings"
    "sort"
)

func DecoratedAnimeTitles(animeInfoList []api.AnimeInfo) []string {
    foundAnimeTitles := make([]string, 0, len(animeInfoList))

	for i, animeInfo := range animeInfoList {
		episodeCount := len(animeInfo.Media.Episodes)
		var b strings.Builder

		fmt.Fprintf(&b, "%d %s", i+1, animeInfo.Names.RU)
		if episodeCount == 0 {
			fmt.Fprintf(&b, " (Нет доступных серий)")
		} else {
			lastDigit := episodeCount % 10
            episodeCase := determineEpisodeCase(lastDigit)
			fmt.Fprintf(&b, " (%d %s)", episodeCount, episodeCase)
		}
		foundAnimeTitles = append(foundAnimeTitles, b.String())
	}
	return foundAnimeTitles
}

// Достаёт название эпизода из значения map
// Если будут дубликаты - крышка
func EpisodesToStrList(episodes map[string]api.Episode) []string{
    episodeIntList := make([]int, 0, len(episodes))
    episodeStrList := make([]string, 0, len(episodes))

    for _, val := range(episodes) {
        episodeIntList = append(episodeIntList, val.Number)
    }
    sort.Ints(episodeIntList)
    
    for _, val := range(episodeIntList) {
        episodeStrList = append(episodeStrList, strconv.Itoa(val))
    }

    return episodeStrList
}

func determineEpisodeCase(lastDigit int) string {
	if lastDigit == 0 || lastDigit >= 5 {
		return "серий"
	}
    if lastDigit < 1 {
        return "серия"
    }
    if lastDigit < 5 {
        return "серии"
    }

	return ""
}

