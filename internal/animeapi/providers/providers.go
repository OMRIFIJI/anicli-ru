package providers

import (
	"strings"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
)

func GetProviders() []string {
	return []string{"animego", "yummyanime", "anilib"}
}

// Задаёт приоритет источников.
// При удалении дубликатов, если число эпизодов совпадает, остается структура источника высшего приоритета.
func getPriorityMap() map[string]int {
	return map[string]int{
		"yummyanime": 2, // Высокий приоритет
		"animego":    1,
		"anilib":     0, // Низкий приоритет
	}
}

// Структура удаляющая дубликаты аниме по названию
type duplicateRemover struct {
	uniqueAnimesMap map[string]models.Anime
	priorityMap     map[string]int
}

func NewDuplicateRemover() *duplicateRemover {
	return &duplicateRemover{
		uniqueAnimesMap: make(map[string]models.Anime),
		priorityMap:     getPriorityMap(),
	}
}

// Отбрасывает дубликаты, оставляя те аниме, у которых больше эпизодов.
// Если количество эпизодов совпадает, то выбирает аниме из более приоритетного
// источника по priorityMap.
func (dp *duplicateRemover) Remove(animes []models.Anime) ([]models.Anime, error) {
	var uniqueAnimes []models.Anime

	for _, anime := range animes {
		dp.tryToInsertAnime(anime)
	}

	for _, anime := range dp.uniqueAnimesMap {
		uniqueAnimes = append(uniqueAnimes, anime)
	}

	return uniqueAnimes, nil
}

// Сравнивать два источника a и b по priorityMap
func (dp *duplicateRemover) isProviderGreater(a, b string) bool {
	return dp.priorityMap[a] > dp.priorityMap[b]
}

func (dp *duplicateRemover) tryToInsertAnime(targetAnime models.Anime) {
	targetTitle := strings.ToLower(strings.TrimSpace(targetAnime.Title))

	if anime, ok := dp.uniqueAnimesMap[targetTitle]; ok {
		if targetAnime.EpCtx.AiredEpCount < anime.EpCtx.AiredEpCount {
			return
		}
		// Если число эпизодов одинаково и новый источник не приоритетнее старого
		if targetAnime.EpCtx.AiredEpCount == anime.EpCtx.AiredEpCount &&
			!dp.isProviderGreater(targetAnime.Provider, anime.Provider) {
			return
		}
	}
	dp.uniqueAnimesMap[targetTitle] = targetAnime
}
