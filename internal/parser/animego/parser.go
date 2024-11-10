package animego

import (
	"anicliru/internal/parser/animego/htmlparser"
	"anicliru/internal/parser/common"
	"anicliru/internal/types"
	"errors"
	"strings"
)

func (a *AnimeGo) FindAnimeByTitle(title string) ([]types.Anime, error) {
	a.init(title)
	
    res, err := a.client.Get(a.url.base + a.url.suff.search + a.title)
	if err != nil {
		return nil, err
	}
    if res.StatusCode != 200 {
        return nil, errors.New("Не получилось соединиться с сервером. Код ошибки: " + res.Status)
    }
	defer res.Body.Close()

	animes, err := htmlparser.GetAnimes(res.Body)
	if err != nil {
		return nil, err
	}
    a.animes = animes

	return a.animes, nil
}

func (a *AnimeGo) init(title string) {
	a.client = common.InitHttpClient()
	a.url = URL{
		base: "https://animego.org/",
		suff: URLSuffixes{
			search: "search/all?q=",
		},
	}
	a.title = strings.TrimSpace(title)
	a.title = strings.ReplaceAll(a.title, " ", "+")
}
