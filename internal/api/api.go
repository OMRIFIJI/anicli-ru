package api

import (
	"anicliru/internal/api/animego"
	"anicliru/internal/api/models"
	"errors"
)

func GetAnimesByTitle(title string) ([]models.Anime, error) {
	client := animego.NewAnimeGoClient()

	animes, err := client.GetAnimesByTitle(title)

	var parseError *models.ParseError
	if err != nil && !errors.As(err, &parseError){
		return animes, err
	}

	return animes, nil
}

func GetEmbedLink(episode *models.Episode) error {
    client := animego.NewAnimeGoClient()
    client.GetEmbedLink(episode)
    
    return nil
}
