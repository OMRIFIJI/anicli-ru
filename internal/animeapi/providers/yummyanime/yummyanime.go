package yummyanime

import (
	"errors"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/providers/yummyanime/parser"
	httpkit "github.com/OMRIFIJI/anicli-ru/internal/httpkit"
	"github.com/OMRIFIJI/anicli-ru/internal/logger"
	"strings"
	"sync"
)

type YummyAnimeClient struct {
	http     *httpkit.HttpClient
	urlBuild urlBuilder
}

func NewYummyAnimeClient(fullDomain string) *YummyAnimeClient {
	y := &YummyAnimeClient{}
	y.http = httpkit.NewHttpClient(
		map[string]string{
			"Accept-Language": "ru-RU,ru;q=0.8",
		},
		httpkit.WithRetries(2),
		httpkit.WithRetryDelay(3),
	)
	y.urlBuild = newUrlBuilder(fullDomain)
	return y
}

func (y *YummyAnimeClient) GetAnimesByTitle(title string) ([]models.Anime, error) {
	foundAnime, err := y.getFoundAnime(title)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
    var mu sync.Mutex

	var animes []models.Anime

	for i, res := range foundAnime {
		wg.Add(1)
		go func(i int, res parser.FoundAnime) {
			defer wg.Done()

			anime := models.Anime{
				Provider:  "yummyanime",
				Id:        res.Id,
				Uname:     res.AnimeUrl,
				Title:     res.Title,
				SearchPos: i,
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
            
            mu.Lock()
            defer mu.Unlock()
			animes = append(animes, anime)
		}(i, res)
	}
	wg.Wait()

	return animes, nil
}

func (y *YummyAnimeClient) getFoundAnime(title string) ([]parser.FoundAnime, error) {
	title = strings.TrimSpace(title)
	url := y.urlBuild.searchByTitle(title)

	res, err := y.http.Get(url)
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

	eps, _ := parser.ParseEpisodes(res.Body)

	if len(eps) == 0 {
		return errors.New("нет доступных ссылок на эпизоды данного аниме")
	}
	anime.EpCtx.Eps = eps

	return nil
}

func (y *YummyAnimeClient) SetEmbedLinks(*models.Anime, *models.Episode) error {
	return nil
}

func (y *YummyAnimeClient) PrepareSavedAnime(anime *models.Anime) error {
	url := y.urlBuild.animeById(anime.Id)

	res, err := y.http.Get(url)
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

	return nil
}
