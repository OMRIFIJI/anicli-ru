package api

import (
	"anicliru/internal/api/animefmt"
	"anicliru/internal/api/animego"
	"anicliru/internal/api/types"
	"errors"
	"sync"
)

type API struct {
	animes []types.Anime
	wg     sync.WaitGroup
}

func (api *API) FindAnimesByTitle(title string) error {
	client := animego.NewAnimeGoClient(
		animego.WithTitle(title),
	)

	animes, err := client.FindAnimesByTitle()
	api.animes = animes

	var parseError *types.ParseError
	if err != nil && !errors.As(err, &parseError){
		return err
	}

	return nil
}

func (api *API) GetAnimeTitlesWrapped() []string {
	wrappedTitles := make([]string, 0, len(api.animes))
	for i, anime := range api.animes {
		wrappedTitle := animefmt.WrapAnimeTitle(anime, i)
		wrappedTitles = append(wrappedTitles, wrappedTitle)
	}
	return wrappedTitles
}
