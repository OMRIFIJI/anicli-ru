package animego

import (
	"anicliru/internal/api/animego/parser"
	"anicliru/internal/api/types"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type AnimeGoURL struct {
	base       string
	searchSuf  string
	animeSuf   string
	episodeSuf string
	playerSuf  string
}

type AnimeGoClient struct {
	client  http.Client
	url     AnimeGoURL
	anime   types.Anime
	title   string
	wg      sync.WaitGroup
	headers map[string]string
}

func NewAnimeGoClient(options ...func(*AnimeGoClient)) *AnimeGoClient {
	animeGo := &AnimeGoClient{}
	animeGo.baseNew()
	for _, o := range options {
		o(animeGo)
	}
	return animeGo
}

func WithTitle(title string) func(*AnimeGoClient) {
	return func(a *AnimeGoClient) {
		a.title = strings.TrimSpace(title)
		a.title = strings.ReplaceAll(a.title, " ", "+")
	}
}

func WithAnime(anime types.Anime) func(*AnimeGoClient) {
	return func(a *AnimeGoClient) {
		a.anime = anime
	}
}

func (a *AnimeGoClient) baseNew() {
	a.client = InitHttpClient()
	a.url = AnimeGoURL{
		base:       "https://animego.org/",
		searchSuf:  "search/anime?q=",
		animeSuf:   "anime/",
		playerSuf:  "player?_allow=true",
		episodeSuf: "series?id=",
	}
	a.headers = map[string]string{
		"User-Agent":       "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0",
		"X-Requested-With": "XMLHttpRequest",
	}
}

func (a *AnimeGoClient) FindAnimesByTitle() ([]types.Anime, error) {
	res, err := a.client.Get(a.url.base + a.url.searchSuf + a.title)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		noConError := types.NoConnectionError{
			Msg: "Не получилось соединиться с сервером. Код ошибки: " + res.Status,
		}
		return nil, &noConError
	}
	defer res.Body.Close()

	animes, err := parser.ParseAnimes(res.Body)
	if err != nil {
		return nil, err
	}

	if len(animes) == 0 {
		notFoundError := types.NotFoundError{
			Msg: "По вашему запросу не удалось ничего найти.",
		}
		return nil, &notFoundError
	}

	for i := 0; i < len(animes); i++ {
		a.wg.Add(1)
		go a.findEpisodeIds(&animes[i])
	}
	a.wg.Wait()

	return animes, nil
}

func (a *AnimeGoClient) findEpisodeIds(anime *types.Anime) {
	defer a.wg.Done()

	animeURL := a.url.base + a.url.animeSuf + anime.Id + "/" + a.url.playerSuf
	req, err := http.NewRequest("GET", animeURL, nil)
	if err != nil {
		return
	}
	for key, val := range a.headers {
		req.Header.Add(key, val)
	}

	res, err := a.client.Do(req)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		return
	}

	epIdMap, lastEpNum, err := parser.ParseEpisodes(res.Body)
	if err != nil {
		return
	}

	// В онгоингах часто сайт может говорить, что доступно на 1 эпизод больше
	if !a.isValidEpisodeId(epIdMap[lastEpNum]) {
		delete(epIdMap, lastEpNum)
	}

	anime.Episodes = make(map[int]types.Episode)
	for key, val := range epIdMap {
		anime.Episodes[key] = types.Episode{
			Id:   val,
			Link: nil,
		}
	}
}

func (a *AnimeGoClient) isValidEpisodeId(episodeId int) bool {
	episodeIdStr := strconv.Itoa(episodeId)
	animeURL := a.url.base + a.url.animeSuf + a.url.episodeSuf + episodeIdStr
	req, err := http.NewRequest("GET", animeURL, nil)
	if err != nil {
		return false
	}
	for key, val := range a.headers {
		req.Header.Add(key, val)
	}

	res, err := a.client.Do(req)
	if err != nil {
		return false
	}
    defer res.Body.Close()

	if res.StatusCode != 200 {
		return false
	}

    isValid := parser.IsValid(res.Body)
    return isValid
}
