package api

import (
	"anicliru/internal/api/animego"
	"anicliru/internal/api/types"
	"errors"
)

type API struct {}

func (api *API) FindAnimesByTitle(title string) ([]types.Anime, error) {
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

