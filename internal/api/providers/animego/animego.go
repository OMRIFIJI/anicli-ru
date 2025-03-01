package animego

import (
	"anicliru/internal/api/models"
	"anicliru/internal/api/providers/animego/parser"
	httpcommon "anicliru/internal/http"
	"anicliru/internal/logger"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type AnimeGoClient struct {
	http     *httpcommon.HttpClient
	urlBuild urlBuilder
}

func NewAnimeGoClient() *AnimeGoClient {
	a := &AnimeGoClient{}
	a.http = httpcommon.NewHttpClient(
		map[string]string{
			"User-Agent":       "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0",
			"Referer":          "https://animego.club",
			"X-Requested-With": "XMLHttpRequest",
			"Accept-Language":  "en-US,en;q=0.5",
		},
		httpcommon.WithRetries(2),
		httpcommon.WithRetryDelay(3),
	)
	a.urlBuild = newUrlBuilder()
	return a
}

func (a *AnimeGoClient) GetAnimesByTitle(title string) ([]models.Anime, error) {
	title = strings.TrimSpace(title)

	url := a.urlBuild.searchByTitle(title)
	res, err := a.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	animes, err := parser.ParseAnimes(res.Body)
	if err != nil {
		logger.ErrorLog.Printf("Ошибка парсинга HTML. %s\n", err)
		return nil, err
	}

	var wg sync.WaitGroup

	for i := 0; i < len(animes); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.getEpsInfo(&animes[i])
		}()
	}
	wg.Wait()

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

	return animesAvailable, nil
}

func (a *AnimeGoClient) getEpsInfo(anime **models.Anime) {
	if err := a.getMediaStatus(*anime); err != nil {
		logger.WarnLog.Printf("Ошибка обработки %s. %s\n", (*anime).Title, err)
		*anime = nil
		return
	}

	// Фильмы могут не иметь информации об id их единственного эпизода
	if (*anime).MediaType == "фильм" {
		if err := a.getFilmRegionBlock(*anime); err != nil {
			logger.WarnLog.Printf("Ошибка обработки %s. %s\n", (*anime).Title, err)
			*anime = nil
		}
		return
	}

	if err := a.getEpIds(*anime); err != nil {
		logger.WarnLog.Printf("Ошибка обработки %s. %s\n", (*anime).Title, err)
		*anime = nil
		return
	}

	// Временная заглушка, надо бы напрямую брать с сайта
	(*anime).EpCtx.AiredEpCount = len((*anime).EpCtx.Eps)
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
		return err
	}

	if isRegionBlock {
		return fmt.Errorf("аниме %s заблокировано на вашей территории", anime.Title)
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
		anime.MediaType = mediaType

		anime.EpCtx.TotalEpCount = 1
		// Сайт не всегда возвращает Id фильмов. В любом случае он не обязателен для работы с ними.
		filmEp := &models.Episode{Id: -1}
		anime.EpCtx.Eps = map[int]*models.Episode{1: filmEp}
		return nil
	}

	if err != nil {
		return fmt.Errorf("ошибка обработки медиа информации %s %s", anime.Title, err)
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

	// В онгоингах часто сайт может говорить, что доступно на 1 эпизод больше, чем есть
	isLastEpValid := true
	if !a.isValidEpId(epIdMap[lastEpNum]) {
		delete(epIdMap, lastEpNum)
	}

	anime.EpCtx.Eps = make(map[int]*models.Episode, len(epIdMap))
	for key, val := range epIdMap {
		anime.EpCtx.Eps[key] = &models.Episode{
			Id: val,
		}
	}

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

func (a *AnimeGoClient) SetAllEmbedLinks(*models.Anime) error {
	return nil
}

func (a *AnimeGoClient) SetEmbedLinks(anime *models.Anime, ep *models.Episode) error {
	var url string
	if anime.MediaType == "фильм" {
		url = a.urlBuild.animeById(anime.Id)
	} else {
		epIdStr := strconv.Itoa(ep.Id)
		url = a.urlBuild.epById(epIdStr)
	}

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
		return errors.New("нет доступных ссылок на выбранный эпизод")
	}
	ep.EmbedLinks = embedLinks

	return nil
}
