package animeapi

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/providers/animego"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/providers/yummyanime"
)

const gistMirrorsUrl = "https://gist.githubusercontent.com/OMRIFIJI/aacb12102b3aff21c37d5273f2b76fa0/raw/anicli-ru-mirrors.csv"

func getProviders() []string {
	return []string{"animego", "yummyanime"}
}

// Задаёт приоритет источников.
// При удалении дубликатов, если число эпизодов совпадает, остается структура источника высшего приоритета.
func getPriorityMap() map[string]int {
	return map[string]int{
		"yummyanime": 1, // Высокий приоритет
		"animego":    0, // Низкий приоритет
	}
}

// Структура удаляющая дубликаты аниме по названию
type duplicateRemover struct {
	uniqueAnimesMap map[string]models.Anime
	priorityMap     map[string]int
}

func newDuplicateRemover() *duplicateRemover {
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

func sortBySearchPos(animes []models.Anime) {
	sort.Slice(animes, func(i, j int) bool {
		return animes[i].SearchPos <= animes[j].SearchPos
	})
}

func SyncedDomainMap() (map[string]string, error) {
	res, err := http.Get(gistMirrorsUrl)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("не удалось связаться с gist github для синхронизации источников")
	}

	resBody := res.Body
	defer resBody.Close()

	reader := csv.NewReader(resBody)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	providersWithParsers := getProviders()

	domainMap := make(map[string]string)
	for _, providerData := range records {
		name, domain := providerData[0], providerData[1]
		if slices.Contains(providersWithParsers, name) {
			domainMap[name] = domain
		}
	}

	return domainMap, nil
}

func newAnimeParserByName(name, fullDomain string) (animeParser, error) {
	switch name {
	case "animego":
		return animego.NewAnimeGoClient(fullDomain), nil
	case "yummyanime":
		return yummyanime.NewYummyAnimeClient(fullDomain), nil
	}
	return nil, fmt.Errorf("парсер %s не существует, проверьте конфиг", name)
}
