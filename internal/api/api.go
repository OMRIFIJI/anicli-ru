package api

import (
	"anicliru/internal/api/animego"
	"anicliru/internal/api/models"
	"anicliru/internal/api/player"
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

func GetEmbedLinks(anime *models.Anime, ep *models.Episode) error {
    client := animego.NewAnimeGoClient()
    client.GetEmbedLinks(anime, ep)
    
    return nil
}

func NewPlayerLinkConverter() *player.PlayerLinkConverter {
	p := player.PlayerLinkConverter{}
	p.SetPlayerHandlers()
	return &p
}
