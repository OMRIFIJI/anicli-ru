package api

import (
	"anicliru/internal/api/animego"
	"anicliru/internal/api/types"
	"errors"
)

func FindAnimesByTitle(title string) ([]types.Anime, error) {
	client := animego.NewAnimeGoClient(
		animego.WithTitle(title),
	)

	animes, err := client.FindAnimesByTitle()

	var parseError *types.ParseError
	if err != nil && !errors.As(err, &parseError){
		return animes, err
	}

	return animes, err
}

func FindEpisodesLinks(anime *types.Anime) error {
    client := animego.NewAnimeGoClient()
    client.FindEpisodesLinks(anime)
    
    return nil
}
