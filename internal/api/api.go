package api

import (
	"anicliru/internal/api/animefmt"
	"anicliru/internal/api/animego"
	"anicliru/internal/api/types"
	"sort"
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

	if err != nil {
		return err
	}

	sort.SliceStable(api.animes, func(i, j int) bool {
		return api.animes[i].IsAvailable && !api.animes[j].IsAvailable
	})

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

