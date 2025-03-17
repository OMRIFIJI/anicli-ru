package parser

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/logger"
)

const nativeHostingDomain = "video1.anilib.me"

func ParseAnimes(r io.Reader) ([]FoundAnime, error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result searchJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func ParseEpCount(r io.Reader) (airedEpCount int, totalEpCount int, err error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return -1, -1, err
	}

	var result animeJson
	if err := json.Unmarshal(in, &result); err != nil {
		return -1, -1, err
	}

	airedEpCount = result.Data.ItemsCount.UploadedCount
	totalEpCount = result.Data.ItemsCount.TotalCount

    // Так надо, но не совсем корректно
    // Частично решает проблему с дробными сериями
	if totalEpCount < airedEpCount {
		totalEpCount = airedEpCount
	}

	return airedEpCount, totalEpCount, nil
}

func ParseEpIds(r io.Reader) (epIdsMap map[int]int, err error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result epMetadataJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, err
	}

	epIdsMap = make(map[int]int)
	for _, epInfo := range result.Data {
		epIdsMap[epInfo.Number] = epInfo.Id
	}

	if len(epIdsMap) == 0 {
		return nil, errors.New("не найден id ни для одного эпизода")
	}

	return epIdsMap, nil
}

func ParseEpisodeEmbed(r io.Reader) (models.EmbedLinks, error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result epVideoJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, err
	}

	embedLinks := make(models.EmbedLinks)

	// Ссылки на видео в json никак не сгруппированы
	for _, res := range result.Data.Players {
		dubName := res.TsType.Label + " " + res.Team.Name

		if _, ok := embedLinks[dubName]; !ok {
			embedLinks[dubName] = make(map[string]string)
		}

		var playerName string
		var link string
		// Случай собственного плеера
		if res.Video != nil {
			playerName = nativeHostingDomain

			if len(res.Video.Quality) == 0 {
                logger.ErrorLog.Println("Len err")
				continue
			}

			// Беру только лучшее качество, все варианты качества с текущей структурой проекта вытягивать некуда
			bestHref := res.Video.Quality[0].Href
			bestQuality := res.Video.Quality[0].Quality
			for _, vid := range res.Video.Quality {
				if vid.Quality > bestQuality {
					bestQuality = vid.Quality
					bestHref = vid.Href
				}
			}

			link = "//" + playerName + "/.%D0%B0s/" + bestHref
		} else {
			if res.Src == "" {
				continue
			}

			u, err := url.Parse(res.Src)
			if err != nil {
				continue
			}

			playerName = u.Hostname()
			link = res.Src

		}
		embedLinks[dubName][playerName] = link
	}

	return embedLinks, nil
}
