package api

import (
	"anicliru/internal/api/animego"
	"anicliru/internal/api/types"
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
	if err != nil {
		return err
	}
	for _, anime := range animes {
		println(anime.Id, anime.Title)
	}
	return nil
}
