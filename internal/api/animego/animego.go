package animego

import (
	"anicliru/internal/api/animego/parser"
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/models"
	httpcommon "anicliru/internal/http"
	"errors"
	"strconv"
	"strings"
	"sync"
)

type animeGoURL struct {
	base       string
	searchSuf  string
	animeSuf   string
	episodeSuf string
	playerSuf  string
}

type AnimeGoClient struct {
	http     *httpcommon.HttpClient
	urlBuild *urlBuilder
	title    string
	headers  map[string]string
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

func (a *AnimeGoClient) baseNew() {
	a.http = httpcommon.NewHttpClient(
		map[string]string{
			"User-Agent":       "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0",
			"X-Requested-With": "XMLHttpRequest",
		},
	)
	a.urlBuild = newUrlBuilder()
}

func (a *AnimeGoClient) FindAnimesByTitle() ([]models.Anime, error) {
	url := a.urlBuild.searchByTitle(a.title)
	res, err := a.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	animes, err := parser.ParseAnimes(res.Body)
	if err != nil {
		apilog.ErrorLog.Printf("Html parse fail. %s\n", err)
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errSlice []error

	for i := 0; i < len(animes); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := a.findMediaInfo(&animes[i]); err != nil {
				mu.Lock()
				errSlice = append(errSlice, err)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	errorComposed := errors.Join(errSlice...)

	var animesAvailable []models.Anime
	for _, anime := range animes {
		if anime != nil {
			animesAvailable = append(animesAvailable, *anime)
		}
	}

	if len(animes) == 0 {
		NotAvailableError := models.NotAvailableError{
			Msg: "По вашему запросу нет доступных аниме.",
		}
		return nil, &NotAvailableError
	}

	return animesAvailable, errorComposed
}

func (a *AnimeGoClient) findMediaInfo(anime **models.Anime) error {
	animeErr := &models.ParseError{
		Msg: "Предупреждение: ошибка при обработке " + (*anime).Title,
	}

	if err := a.findEpisodeCount(*anime); err != nil {
		*anime = nil
		return animeErr
	}

	// Фильмы могут не иметь информации об их id
	if (*anime).TotalEpCount == 1 {
		if err := a.findFilmRegionBlock(*anime); err != nil {
			*anime = nil
			return animeErr
		}
		return nil
	}

	if err := a.findEpisodeIds(*anime); err != nil {
		*anime = nil

		var blockError *models.RegionBlockError
		if !errors.As(err, &blockError) {
			return animeErr
		}
	}

	return nil
}

func (a *AnimeGoClient) findFilmRegionBlock(anime *models.Anime) error {
	url := a.urlBuild.animeById(anime.Id)
	res, err := a.http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	isRegionBlock, err := parser.ParseFilmRegionBlock(res.Body)
	if err != nil {
		apilog.ErrorLog.Printf("Parse error. %s\n", err)
		return err
	}

	if isRegionBlock {
		err := &models.RegionBlockError{
			Msg: "Не доступно на территории РФ",
		}
		return err
	}

	return nil
}

func (a *AnimeGoClient) findEpisodeCount(anime *models.Anime) error {
	url := a.urlBuild.animeByUname(anime.Uname)
	res, err := a.http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	episodeCount, err := parser.ParseEpisodeCount(res.Body)
	if err != nil {
		apilog.ErrorLog.Printf("Parse error. %s\n", err)
		return err
	}
	anime.TotalEpCount = episodeCount

	return nil
}

func (a *AnimeGoClient) findEpisodeIds(anime *models.Anime) error {
	url := a.urlBuild.animeById(anime.Id)
	res, err := a.http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	epIdMap, lastEpNum, err := parser.ParseSeriesEpisodes(res.Body)
	if err != nil {
		return err
	}

	// В онгоингах часто сайт может говорить, что доступно на 1 эпизод больше, чем есть
	if !a.isValidEpisodeId(epIdMap[lastEpNum]) {
		delete(epIdMap, lastEpNum)
	}

	anime.Episodes = make(map[int]*models.Episode)
	for key, val := range epIdMap {
		anime.Episodes[key] = &models.Episode{
			Id: val,
		}
	}
	return nil
}

func (a *AnimeGoClient) isValidEpisodeId(episodeId int) bool {
	episodeIdStr := strconv.Itoa(episodeId)
	url := a.urlBuild.episodeById(episodeIdStr)

	res, err := a.http.Get(url)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	isValid := parser.IsValid(res.Body)
	return isValid
}

func (a *AnimeGoClient) FindEpisodesLinks(anime *models.Anime) error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	hasFoundEpisode := false

	for key, val := range anime.Episodes {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := a.findEpisodeLinks(key, val)
			if err == nil {
				mu.Lock()
				hasFoundEpisode = true
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if !hasFoundEpisode {
		err := &models.NotFoundError{
			Msg: "Не удалось найти ни один эпизод.",
		}
		return err
	}

	return nil
}

func (a *AnimeGoClient) findEpisodeLinks(episodeNum int, episode *models.Episode) error {
	episodeIdStr := strconv.Itoa(episodeNum)
	url := a.urlBuild.episodeById(episodeIdStr)

	res, err := a.http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	playerLinks, err := parser.ParsePlayerLinks(res.Body)
	episode.EmbedLink = playerLinks

	return nil
}

