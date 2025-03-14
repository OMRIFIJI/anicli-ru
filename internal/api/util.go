package api

import (
	"anicliru/internal/api/models"
	config "anicliru/internal/app/cfg"
	"anicliru/internal/db"
	httpcommon "anicliru/internal/http"
	"encoding/csv"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
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

func sortBySearchPos(animes []models.Anime) {
	sort.Slice(animes, func(i, j int) bool {
		return animes[i].SearchPos <= animes[j].SearchPos
	})
}

func removeUnreachableProviders(providers map[string]string, dialer *httpcommon.Dialer) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for key, domain := range providers {
		wg.Add(1)
		go func(string, string) {
			defer wg.Done()
			url := "https://" + domain

			_, err := dialer.Ping(url)
			if err != nil {
				// logger.ErrorLog.Println("Нет связи с источником %s", domain)
				mu.Lock()
				defer mu.Unlock()
				delete(providers, key)
			}
		}(key, domain)
	}
	wg.Wait()
}

func isTimeToSync(cfg *config.Config, dbh *db.DBHandler, currentTime time.Time) bool {
	// Пустая строка - синхронизация отключена
	if len(cfg.Players.SyncInterval) == 0 {
		return false
	}

	lastSyncTime, err := dbh.GetLastSyncTime()
	if err != nil {
		return true
	}
	diff := currentTime.Sub(*lastSyncTime)
	days := int(diff.Hours() / 24)

	syncIntervalStr := cfg.Players.SyncInterval
	syncInterval, err := strconv.Atoi(syncIntervalStr[:len(syncIntervalStr)-1])

	return days >= syncInterval
}

func newDomainMap(cfg *config.Config) (map[string]string, error) {
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

	domainMap := make(map[string]string)

	for _, providerData := range records {
		name, domain := providerData[0], providerData[1]
		domainMap[name] = domain
	}

	return domainMap, nil
}
