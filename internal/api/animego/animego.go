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

type AnimeGoClient struct {
	http     *httpcommon.HttpClient
	urlBuild *urlBuilder
}

func NewAnimeGoClient() *AnimeGoClient {
	a := &AnimeGoClient{}
	a.http = httpcommon.NewHttpClient(
		map[string]string{
			"User-Agent":       "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0",
			"X-Requested-With": "XMLHttpRequest",
		},
		httpcommon.WithRetries(2),
		httpcommon.WithRetryDelay(10),
	)
	a.urlBuild = newUrlBuilder()
	return a
}

func (a *AnimeGoClient) GetAnimesByTitle(title string) ([]models.Anime, error) {
	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, " ", "+")

	url := a.urlBuild.searchByTitle(title)
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

			if err := a.getEpInfo(&animes[i]); err != nil {
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

func (a *AnimeGoClient) getEpInfo(anime **models.Anime) error {
	animeErr := &models.ParseError{
		Msg: "Предупреждение: ошибка при обработке " + (*anime).Title,
	}

	if err := a.getMediaStatus(*anime); err != nil {
		*anime = nil
		return animeErr
	}

	// Фильмы могут не иметь информации об id их единственного эпизода
	if (*anime).MediaType == "фильм" {
		if err := a.getFilmRegionBlock(*anime); err != nil {
			*anime = nil
			return animeErr
		}
		return nil
	}

	if err := a.getEpIds(*anime); err != nil {
		var blockError *models.RegionBlockError
		if !errors.As(err, &blockError) {
			apilog.ErrorLog.Printf("Ошибка обработки %s %s\n", (*anime).Title, err)
			*anime = nil
			return animeErr
		}
		*anime = nil
	}

	return nil
}

func (a *AnimeGoClient) getFilmRegionBlock(anime *models.Anime) error {
	url := a.urlBuild.animeById(anime.Id)
	res, err := a.http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	isRegionBlock, err := parser.ParseFilmRegionBlock(res.Body)
	if err != nil {
		apilog.ErrorLog.Printf("Ошибка обработки блокировки фильма %s %s\n", anime.Title, err)
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

func (a *AnimeGoClient) getMediaStatus(anime *models.Anime) error {
	url := a.urlBuild.animeByUname(anime.Uname)
	res, err := a.http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	epCount, mediaType, err := parser.ParseMediaStatus(res.Body)
	if mediaType == "фильм" {
		anime.EpCtx.TotalEpCount = 1
		anime.MediaType = mediaType

		filmEp := &models.Episode{Id: models.FilmEpisodeId}
		anime.EpCtx.Eps = map[int]*models.Episode{1: filmEp}
		return nil
	}

	if err != nil {
		apilog.ErrorLog.Printf("Ошибка обработки медиа информации %s %s\n", anime.Title, err)
		return err
	}

	anime.EpCtx.TotalEpCount = epCount
	anime.MediaType = mediaType

	return nil
}

func (a *AnimeGoClient) getEpIds(anime *models.Anime) error {
	url := a.urlBuild.animeById(anime.Id)
	res, err := a.http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	epIdMap, lastEpNum, err := parser.ParseEpIds(res.Body)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	// В онгоингах часто сайт может говорить, что доступно на 1 эпизод больше, чем есть
	isLastEpValid := true
	go func() {
		defer wg.Done()
		isLastEpValid = a.isValidEpId(epIdMap[lastEpNum])
	}()
	if !a.isValidEpId(epIdMap[lastEpNum]) {
		delete(epIdMap, lastEpNum)
	}

	anime.EpCtx.Eps = make(map[int]*models.Episode)
	for key, val := range epIdMap {
		anime.EpCtx.Eps[key] = &models.Episode{
			Id: val,
		}
	}

	wg.Wait()
	if !isLastEpValid {
		delete(anime.EpCtx.Eps, lastEpNum)
	}

	return nil
}

func (a *AnimeGoClient) isValidEpId(epId int) bool {
	epIdStr := strconv.Itoa(epId)
	url := a.urlBuild.epById(epIdStr)

	res, err := a.http.Get(url)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	isValid := parser.IsValid(res.Body)
	return isValid
}

func (a *AnimeGoClient) GetEmbedLinks(ep *models.Episode) error {
	epIdStr := strconv.Itoa(ep.Id)
	url := a.urlBuild.epById(epIdStr)

	res, err := a.http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()


	embedLinks, err := parser.ParseEmbedLinks(res.Body)
    if err != nil {
        return err
    }
    if len(embedLinks) == 0 {
        return errors.New("Нет доступных ссылок на выбранный эпизод.")
    }
	ep.EmbedLinks = embedLinks

	return nil
}
