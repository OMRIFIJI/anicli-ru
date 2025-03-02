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

func NewAnimeParserByName(name string) (animeParser, error) {
	switch name {
	case "animego":
		return animego.NewAnimeGoClient(), nil
	case "yummyanime":
		return yummyanime.NewYummyAnimeClient(), nil
	}
	return nil, fmt.Errorf("парсер %s не существует", name)
}

func NewAnimeAPI(animeParserNames []string) (*AnimeAPI, error) {
	a := AnimeAPI{}
	a.animeParsers = make(map[string]animeParser)
	for _, name := range animeParserNames {
		animeParser, err := NewAnimeParserByName(name)
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
			animes = append(animes, parsedAnimes...)
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(animes) == 0 {
		return nil, errors.New("По вашему запросу ничего не найдено.")
	}

	animes, err := dropAnimeDuplicates(animes)
	if err != nil {
		return nil, err
	}

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
		return fmt.Errorf("парсер %s не доступен", anime.Provider)
	}
	if err := client.PrepareSavedAnime(anime); err != nil {
		return err
	}
	return nil
}

func NewPlayerLinkConverter() *player.PlayerLinkConverter {
	p := player.PlayerLinkConverter{}
	p.SetPlayerHandlers()
	return &p
}
