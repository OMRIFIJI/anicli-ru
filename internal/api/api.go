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
    api.animes = animes

    for _, anime := range animes {
        println()
		println(anime.Id, anime.Uname, anime.Title, anime.TotalEpCount)
	}
	if err != nil {
		return err
	}
	

	return nil
}
