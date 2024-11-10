package api

import (
	"anicliru/internal/api/animego"
	"anicliru/internal/api/types"
	"errors"
)

type API struct {
	parser      types.Parser
	animesFound []types.Anime
}

func (api *API) FindAnimeByTitle(title string) error {
	api.parser = &animego.AnimeGo{}

	animesFound, err := api.parser.FindAnimeByTitle(title)
	if err != nil {
		return err
	}

	// на == nil не менять
	if len(animesFound) == 0 {
		return errors.New("По вашему запросу ничего не нашлось.")
	}
	api.animesFound = animesFound

	for _, anime := range api.animesFound {
		print("Название аниме: " + anime.Title + ", ")
		print("ссылка: " + anime.Link + ".\n")
	}

	return err
}
