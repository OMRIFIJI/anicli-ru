package api

import (
	"anicliru/internal/api/models"
	"strings"
)

func tryToAppendTitle(targetAnime models.Anime, uniqueAnimesMap map[string]models.Anime) {
	targetTitle := strings.ToLower(strings.TrimSpace(targetAnime.Title))

	if anime, ok := uniqueAnimesMap[targetTitle]; ok &&
		targetAnime.EpCtx.AiredEpCount <= anime.EpCtx.AiredEpCount {
		return
	}
	uniqueAnimesMap[targetTitle] = targetAnime
}

// Отбрасывает дубликаты, оставляя те аниме, у которых больше эпизодов
func dropAnimeDuplicates(animes []models.Anime) ([]models.Anime, error) {
	uniqueAnimesMap := make(map[string]models.Anime)
	var uniqueAnimes []models.Anime

	for _, anime := range animes {
		tryToAppendTitle(anime, uniqueAnimesMap)
	}

	for _, anime := range uniqueAnimesMap {
		uniqueAnimes = append(uniqueAnimes, anime)
	}

	return uniqueAnimes, nil
}
