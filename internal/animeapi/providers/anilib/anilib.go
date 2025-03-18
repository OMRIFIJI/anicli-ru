package anilib

import (
	"strconv"
	"strings"
	"sync"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/providers/anilib/parser"
	httpkit "github.com/OMRIFIJI/anicli-ru/internal/httpkit"
	"github.com/OMRIFIJI/anicli-ru/internal/logger"
)

// Источник может увлечься и отправить 50-60 аниме, а со стороны сервера limit в api не работает
const animesMaxLen = 10

type AniLibClient struct {
	http     *httpkit.HttpClient
	urlBuild urlBuilder
}

func NewAniLibClient(fullDomain string) *AniLibClient {
	a := &AniLibClient{}
	a.http = httpkit.NewHttpClient(
		map[string]string{
			"Accept-Language": "ru-RU,ru;q=0.8",
		},
		httpkit.WithRetries(2),
		httpkit.WithRetryDelay(3),
	)
	a.urlBuild = newUrlBuilder(fullDomain)
	return a
}

func (a *AniLibClient) GetAnimesByTitle(title string) ([]models.Anime, error) {
	foundAnime, err := a.getFoundAnime(title)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	if len(foundAnime) > animesMaxLen {
		foundAnime = foundAnime[:animesMaxLen]
	}

    var mu sync.Mutex
	var animes []models.Anime

	for i, data := range foundAnime {
		wg.Add(1)
		go func(i int, data parser.FoundAnime) {
			defer wg.Done()

			anime := models.Anime{
				Provider:  "anilib",
				Id:        data.Id,
				Uname:     data.Slug,
				Title:     data.Title,
				SearchPos: i,
			}

			slugUrl := buildSlugUrl(anime.Id, anime.Uname)

			typeName := strings.ToLower(data.Type.Label)
			if strings.Contains(typeName, "фильм") {
				typeName = "фильм"

				anime.MediaType = typeName
				anime.EpCtx = models.EpisodesContext{
					TotalEpCount: 1,
					AiredEpCount: 1,
				}
			} else {
				airedEpCount, totalEpCount, err := a.getEpCount(slugUrl)
				// Может быть всё равно выводить?
				if err != nil {
					return
				}

				anime.MediaType = typeName
				anime.EpCtx = models.EpisodesContext{
					TotalEpCount: totalEpCount,
					AiredEpCount: airedEpCount,
				}
			}

			// Получаем id эпизодов
			eps, err := a.getEpsWithId(slugUrl)
			if err != nil {
				return
			}
			anime.EpCtx.Eps = eps

            mu.Lock()
            defer mu.Unlock()
			animes = append(animes, anime)
		}(i, data)
	}
	wg.Wait()

	return animes, nil
}

func (a *AniLibClient) getFoundAnime(title string) ([]parser.FoundAnime, error) {
	title = strings.TrimSpace(title)
	url := a.urlBuild.searchByTitle(title)

	res, err := a.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	foundAnime, err := parser.ParseAnimes(res.Body)
	if err != nil {
		logger.ErrorLog.Printf("Ошибка парсинга. %s\n", err)
		return nil, err
	}

	return foundAnime, nil
}

func (a *AniLibClient) getEpCount(slugUrl string) (airedEpCount int, totalEpCount int, err error) {
	url := a.urlBuild.animeBySlugUrl(slugUrl)

	res, err := a.http.Get(url)
	if err != nil {
		return -1, -1, err
	}
	defer res.Body.Close()

	airedEpCount, totalEpCount, err = parser.ParseEpCount(res.Body)
	if err != nil {
		return -1, -1, err
	}

	return airedEpCount, totalEpCount, nil
}

func (a *AniLibClient) getEpsWithId(slugUrl string) (eps map[int]models.Episode, err error) {
	url := a.urlBuild.epsIdBySlugUrl(slugUrl)

	res, err := a.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	epIdsMap, err := parser.ParseEpIds(res.Body)
	if err != nil {
		return nil, err
	}

	eps = make(map[int]models.Episode)
	for epNum, epId := range epIdsMap {
		eps[epNum] = models.Episode{Id: epId}
	}

	return eps, nil
}

func (a *AniLibClient) SetAllEmbedLinks(*models.Anime) error {
	return nil
}

func (a *AniLibClient) SetEmbedLinks(anime *models.Anime, ep *models.Episode) error {
	url := a.urlBuild.embedByEpId(ep.Id)
	res, err := a.http.Get(url)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	embedLinks, err := parser.ParseEpisodeEmbed(res.Body)
	if err != nil {
		return err
	}
	ep.EmbedLinks = embedLinks

	return nil
}

func (a *AniLibClient) PrepareSavedAnime(anime *models.Anime) error {
	slugUrl := buildSlugUrl(anime.Id, anime.Uname)
	url := a.urlBuild.animeBySlugUrl(slugUrl)

	res, err := a.http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	airedEpCount, totalEpCount, err := parser.ParseEpCount(res.Body)
	if err != nil {
		return err
	}
	anime.EpCtx.AiredEpCount = airedEpCount
	anime.EpCtx.TotalEpCount = totalEpCount

	// собираем id эпизодов
	eps, err := a.getEpsWithId(slugUrl)
	if err != nil {
		return err
	}
	anime.EpCtx.Eps = eps

	return nil
}

func buildSlugUrl(animeId int, slug string) string {
	return strconv.Itoa(animeId) + "--" + slug
}
