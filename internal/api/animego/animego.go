package animego

import (
	"anicliru/internal/api/animego/parser"
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/types"
	"errors"
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
	defer res.Body.Close()

	if res.StatusCode != 200 {
		noConError := types.HttpError{
			Msg: "Не получилось соединиться с сервером. Код ошибки: " + res.Status,
		}
		return nil, &noConError
	}

	animes, err := parser.ParseAnimes(res.Body)
	if err != nil {
		apilog.ErrorLog.Printf("Html parse fail. %s\n", err)
		return nil, err
	}

	if len(animes) == 0 {
		notFoundError := types.NotFoundError{
			Msg: "По вашему запросу не удалось ничего найти.",
		}
		return nil, &notFoundError
	}

	errChan := make(chan error, len(animes))
	for i := 0; i < len(animes); i++ {
		a.wg.Add(1)
		go a.findMediaInfo(&animes[i], errChan)
	}

	go func() {
		a.wg.Wait()
		close(errChan)
	}()

	errMsg := ""
	for err := range errChan {
		if err != nil {
			errMsg += err.Error() + "\n"
		}
	}

	if errMsg == "" {
		return animes, nil
	} else {
		composedErr := &types.AnimeError{
			Msg: errMsg,
		}
		return animes, composedErr
	}
}

func (a *AnimeGoClient) findMediaInfo(anime *types.Anime, errChan chan error) {
	defer a.wg.Done()

	animeErr := &types.AnimeError{
		Msg: "Предупреждение: ошибка при обработке " + (*anime).Title,
	}

	if err := a.findMediaStatus(anime); err != nil {
		errChan <- animeErr
		anime.IsAvailable = false
		return
	}

	// Фильмы могут не иметь информации об их id
	if anime.IsFilm {
		err := a.findFilmRegionBlock(anime)
		anime.IsAvailable = err == nil
		return
	}

	// Для сериалов соберем id
	if err := a.findEpisodeIds(anime); err != nil {
		var blockError *types.RegionBlockError
		if errors.As(err, &blockError) {
			anime.IsRegionBlock = true
		} else {
			errChan <- animeErr
		}
		anime.IsAvailable = false
		return
	}

	anime.IsRegionBlock = false
	anime.IsAvailable = true
}

func (a *AnimeGoClient) findFilmRegionBlock(anime *types.Anime) (err error) {
	animeURL := a.url.base + a.url.animeSuf + anime.Id + "/" + a.url.playerSuf
	req, err := http.NewRequest("GET", animeURL, nil)
	if err != nil {
		return err
	}
	for key, val := range a.headers {
		req.Header.Add(key, val)
	}

	res, err := a.client.Do(req)
	if err != nil {
		apilog.ErrorLog.Printf("Http error. %s\n", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return err
	}

	isRegionBlock, err := parser.ParseFilmRegionBlock(res.Body)
	if err != nil {
		apilog.ErrorLog.Printf("Parse error. %s\n", err)
		return err
	}
	anime.IsRegionBlock = isRegionBlock
	if isRegionBlock {
		anime.IsAvailable = false
	}
	return nil
}

func (a *AnimeGoClient) findMediaStatus(anime *types.Anime) error {
	res, err := a.client.Get(a.url.base + a.url.animeSuf + anime.Uname)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return err
	}

	episodeCount, isFilm, err := parser.ParseMediaStatus(res.Body)
	if err != nil {
		apilog.ErrorLog.Printf("Parse error. %s\n", err)
		return err
	}

	anime.IsFilm = isFilm
	anime.TotalEpCount = episodeCount

	return nil
}

func (a *AnimeGoClient) findEpisodeIds(anime *types.Anime) error {
	animeURL := a.url.base + a.url.animeSuf + anime.Id + "/" + a.url.playerSuf
	req, err := http.NewRequest("GET", animeURL, nil)
	if err != nil {
		return err
	}
	for key, val := range a.headers {
		req.Header.Add(key, val)
	}

	res, err := a.client.Do(req)
	if err != nil {
		apilog.ErrorLog.Printf("Http error. %s\n", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		apilog.ErrorLog.Printf("Http error. %s\n", err)
		return err
	}

	epIdMap, lastEpNum, err := parser.ParseSeriesEpisodes(res.Body)
	if err != nil {
		apilog.ErrorLog.Printf("Parse error. %s\n", err)
		return err
	}

	// В онгоингах часто сайт может говорить, что доступно на 1 эпизод больше, чем есть
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
	return nil
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
		apilog.ErrorLog.Printf("Http error. %s\n", err)
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		apilog.ErrorLog.Printf("Http error. %s\n", err)
		return false
	}

	isValid := parser.IsValid(res.Body)
	return isValid
}
