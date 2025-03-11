package api

import (
	"anicliru/internal/api/models"
	"anicliru/internal/api/player"
	"anicliru/internal/api/providers/animego"
	"anicliru/internal/api/providers/yummyanime"
	"anicliru/internal/logger"
	"errors"
	"fmt"
	"sync"
)

type AnimeAPI struct {
	animeParsers map[string]animeParser
}

// TODO: исправить логику с embed'ами.
type animeParser interface {
	GetAnimesByTitle(string) ([]models.Anime, error)
	SetEmbedLinks(*models.Anime, *models.Episode) error
	SetAllEmbedLinks(*models.Anime) error
	// Дозаполняет структуру аниме из сохраненных перед вычислением embed'ов
	PrepareSavedAnime(anime *models.Anime) error
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

func NewAnimeAPI(Providers map[string]string) (*AnimeAPI, error) {
	a := AnimeAPI{}
	a.animeParsers = make(map[string]animeParser)
	for name, fullDomain := range Providers {
		animeParser, err := NewAnimeParserByName(name, fullDomain)
		if err != nil {
			return nil, err
		}
		a.animeParsers[name] = animeParser
	}

	if len(a.animeParsers) == 0 {
		return nil, errors.New("не удалось найти ни один парсер в конфиге")
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

func NewPlayerLinkConverter() *player.PlayerLinkConverter {
	return player.NewPlayerLinkConverter()
}
