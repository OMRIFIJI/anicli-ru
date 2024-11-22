package api

import (
	"anicliru/internal/api/animego"
	"anicliru/internal/api/models"
	"errors"
)

func FindAnimesByTitle(title string) ([]models.Anime, error) {
	client := animego.NewAnimeGoClient(
		animego.WithTitle(title),
	)

	animes, err := client.FindAnimesByTitle()

	var parseError *models.ParseError
	if err != nil && !errors.As(err, &parseError){
		return animes, err
	}

	return animes, nil
}

func FindEpisodesLinks(anime *models.Anime) error {
    client := animego.NewAnimeGoClient()
    client.FindEpisodesLinks(anime)
    
    return nil
}
