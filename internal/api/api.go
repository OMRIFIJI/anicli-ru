package api

import (
	"anicliru/internal/api/models"
	"anicliru/internal/api/player"
	"anicliru/internal/api/player/common"
	"anicliru/internal/api/providers/animego"
	"anicliru/internal/api/providers/yummyanime"
	config "anicliru/internal/app/cfg"
	"anicliru/internal/db"
	httpcommon "anicliru/internal/http"
	"anicliru/internal/logger"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type AnimeAPI struct {
	animeParsers map[string]animeParser
	Converter    *player.PlayerLinkConverter
}

// TODO: исправить логику с embed'ами.
type animeParser interface {
	GetAnimesByTitle(string) ([]models.Anime, error)
	SetEmbedLinks(*models.Anime, *models.Episode) error
	SetAllEmbedLinks(*models.Anime) error
	// Дозаполняет структуру аниме из сохраненных перед вычислением embed'ов
	PrepareSavedAnime(anime *models.Anime) error
}

func GetProvidersState(providers map[string]string) string {
	dialer := httpcommon.NewDialer()
	var wg sync.WaitGroup
	var mu sync.Mutex

	var b strings.Builder
	for key, provider := range providers {
		key := key
        provider := provider
		providerLink := "http://" + provider

		wg.Add(1)
		go func() {
			defer wg.Done()

			if _, err := dialer.Ping(providerLink); err != nil {
				mu.Lock()
				defer mu.Unlock()
				fmt.Fprintf(&b, "Источник %s - %s не доступен\n", key, provider)
			} else {
				mu.Lock()
				defer mu.Unlock()
				fmt.Fprintf(&b, "Источник %s - %s доступен\n", key, provider)
			}
		}()
	}

	wg.Wait()

    return b.String()
}

func NewAnimeParserByName(name, fullDomain string) (animeParser, error) {
	switch name {
	case "animego":
		return animego.NewAnimeGoClient(fullDomain), nil
	case "yummyanime":
		return yummyanime.NewYummyAnimeClient(fullDomain), nil
	}
	return nil, fmt.Errorf("парсер %s не существует, проверьте конфиг", name)
}

func NewAnimeAPI(cfg *config.Config, dbh *db.DBHandler) (*AnimeAPI, error) {
	providers := cfg.Providers
	animeParsers := make(map[string]animeParser)

	for name, fullDomain := range providers {
		animeParser, err := NewAnimeParserByName(name, fullDomain)
		if err != nil {
			return nil, err
		}
		animeParsers[name] = animeParser
	}

	if len(animeParsers) == 0 {
		return nil, errors.New("не удалось найти ни один парсер в конфиге")
	}

	var converter *player.PlayerLinkConverter
	var err error
	currentTime := time.Now().UTC()

	if isTimeToSync(cfg, dbh, currentTime) {
		converter, err = player.NewPlayerLinkConverter(player.WithSync(dbh))
		if err != nil {
			return nil, err
		}

		domainMap := common.NewPlayerDomainMap()
		var newDomains []string
		for origin := range converter.Handlers {
			newDomains = append(newDomains, domainMap[origin])
		}

		cfg.Players.Domains = newDomains
		cfg.Write()
		dbh.UpdateLastSyncTime(currentTime)
	} else {
		converter, err = player.NewPlayerLinkConverter(player.FromConfig(cfg))
		if err != nil {
			return nil, err
		}
	}

	a := AnimeAPI{
		animeParsers: animeParsers,
		Converter:    converter,
	}
	return &a, nil
}

func (a *AnimeAPI) GetAnimesByTitle(title string) ([]models.Anime, error) {
	var animes []models.Anime
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, client := range a.animeParsers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			parsedAnimes, err := client.GetAnimesByTitle(title)
			if err != nil {
				logger.ErrorLog.Println(err)
			}

			mu.Lock()
			defer mu.Unlock()
			animes = append(animes, parsedAnimes...)
		}()
	}
	wg.Wait()

	if len(animes) == 0 {
		return nil, errors.New("по вашему запросу ничего не найдено")
	}

	animes, err := dropAnimeDuplicates(animes)
	if err != nil {
		return nil, err
	}
	sortBySearchPos(animes)

	return animes, nil
}

// Пытается установить все embed если это возможно.
func (a *AnimeAPI) SetAllEmbedLinks(anime *models.Anime) error {
	client := a.animeParsers[anime.Provider]
	err := client.SetAllEmbedLinks(anime)
	if err != nil {
		return err
	}

	return nil
}

func (a *AnimeAPI) SetEmbedLinks(anime *models.Anime, ep *models.Episode) error {
	client := a.animeParsers[anime.Provider]

	err := client.SetEmbedLinks(anime, ep)
	if err != nil {
		return err
	}

	return nil
}

func (a *AnimeAPI) PrepareSavedAnime(anime *models.Anime) error {
	client, ok := a.animeParsers[anime.Provider]
	if !ok {
		// Зануляем источник, если его больше нет в конфиге
		anime.Provider = ""
		return fmt.Errorf("парсер %s не доступен, проверьте конфиг", anime.Provider)
	}
	if err := client.PrepareSavedAnime(anime); err != nil {
		// Зануляем источник, если больше нет ответа
		anime.Provider = ""
		return err
	}
	return nil
}
