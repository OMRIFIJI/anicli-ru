package api

import (
	"anicliru/internal/api/models"
	"anicliru/internal/api/player"
	"anicliru/internal/api/providers/animego"
	"anicliru/internal/api/providers/yummyanime"
	"errors"
	"fmt"
)

type AnimeAPI struct {
	animeParsers []animeParser
}

type animeParser interface {
	GetAnimesByTitle(string) ([]models.Anime, error)
	// SetEmbedLinks(*models.Anime, *models.Episode) error
}

func getAnimeParserByName(name string) (animeParser, error) {
	switch name {
	case "animego":
		return animego.NewAnimeGoClient(), nil
	case "yummyanime":
		return yummyanime.NewYummyAnimeClient(), nil
	}
	return nil, fmt.Errorf("Парсер %s не существует.", name)
}

func NewAnimeAPI(animeParserNames []string) (*AnimeAPI, error) {
	a := AnimeAPI{}
	for _, name := range animeParserNames {
		animeParser, err := getAnimeParserByName(name)
		if err != nil {
			return nil, err
		}
		a.animeParsers = append(a.animeParsers, animeParser)
	}

	return &a, nil
}

func (a *AnimeAPI) GetAnimesByTitle(title string) ([]models.Anime, error) {
    var animes []models.Anime

	for _, client := range a.animeParsers {
		parsedAnimes, err := client.GetAnimesByTitle(title)
		if err != nil {
			return nil, err
		}

        animes = append(animes, parsedAnimes...)
	}

    if len(animes) == 0 {
        return nil, errors.New("По вашему запросу ничего не найдено.")
    }
    
	return animes, nil
}

func SetEmbedLinks(anime *models.Anime, ep *models.Episode) error {
	client := animego.NewAnimeGoClient()
	client.SetEmbedLinks(anime, ep)

	return nil
}

func NewPlayerLinkConverter() *player.PlayerLinkConverter {
	p := player.PlayerLinkConverter{}
	p.SetPlayerHandlers()
	return &p
}
