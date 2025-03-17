package animeapi

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/providers/animego"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/providers/yummyanime"
	"net/http"
	"slices"
	"sort"
	"strings"
)

const gistMirrorsUrl = "https://gist.githubusercontent.com/OMRIFIJI/aacb12102b3aff21c37d5273f2b76fa0/raw/anicli-ru-mirrors.csv"

func getProviders() []string {
	return []string{"animego", "yummyanime"}
}

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
