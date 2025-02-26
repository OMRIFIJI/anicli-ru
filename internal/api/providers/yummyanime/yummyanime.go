package yummyanime

import (
	"anicliru/internal/api/models"
	"anicliru/internal/api/providers/yummyanime/parser"
	httpcommon "anicliru/internal/http"
	"anicliru/internal/logger"
	"errors"
	"strings"
	"sync"
)

type YummyAnimeClient struct {
	http     *httpcommon.HttpClient
	urlBuild urlBuilder
}

func NewYummyAnimeClient() *YummyAnimeClient {
	y := &YummyAnimeClient{}
	y.http = httpcommon.NewHttpClient(
		map[string]string{
			"Accept-Language": "ru-RU,ru;q=0.8",
		},
		httpcommon.WithRetries(2),
		httpcommon.WithRetryDelay(3),
	)
	y.urlBuild = newUrlBuilder()
	return y
}

func (y *YummyAnimeClient) GetAnimesByTitle(title string) ([]models.Anime, error) {
	searchJson, err := y.getSearchJson(title)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	animes := make([]models.Anime, len(searchJson.Response))

	for i, res := range searchJson.Response {
		wg.Add(1)
		go func() {
			defer wg.Done()

			anime := models.Anime{
				Provider: "yummyanime",
				Id:       res.Id,
				Uname:    res.AnimeUrl,
				Title:    res.Title,
			}

			typeName := strings.ToLower(res.Type.Name)
			if strings.Contains(typeName, "фильм") {
				typeName = "фильм"

				anime.MediaType = typeName
				anime.EpCtx = models.EpisodesContext{
					TotalEpCount: 1,
					AiredEpCount: 1,
				}
			} else {
				airedEpCount, totalEpCount, err := y.getEpCount(res.Id)
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

			animes[i] = anime
		}()
	}
	wg.Wait()

	return animes, nil
}

func (y *YummyAnimeClient) getSearchJson(title string) (*parser.SearchJson, error) {
	title = strings.TrimSpace(title)
	url := y.urlBuild.searchByTitle(title)

	res, err := y.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	searchJson, err := parser.ParseAnimes(res.Body)
	if err != nil {
		logger.ErrorLog.Printf("Ошибка парсинга HTML. %s\n", err)
		return nil, err
	}

	return searchJson, nil
}

func (y *YummyAnimeClient) getEpCount(animeId int) (airedEpCount int, totalEpCount int, err error) {
	url := y.urlBuild.animeById(animeId)

	res, err := y.http.Get(url)
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

func (y *YummyAnimeClient) SetAllEmbedLinks(anime *models.Anime) error {
	url := y.urlBuild.embedByAnimeId(anime.Id)
	res, err := y.http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Игнорируем входной эпизод и заполняем все эпизоды
	eps, _ := parser.ParseEpisodes(res.Body)

	logger.WarnLog.Printf("Found %d eps\n", len(eps))
	if len(eps) == 0 {
		return errors.New("нет доступных ссылок на эпизоды данного аниме")
	}
	anime.EpCtx.Eps = eps

	return nil
}

func (y *YummyAnimeClient) SetEmbedLinks(*models.Anime, *models.Episode) error {
	return nil
}
