package parser

import (
	"encoding/json"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"io"
	"net/url"
	"strconv"
)

func ParseAnimes(r io.Reader) ([]FoundAnime, error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result searchJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, err
	}

	return result.Response, nil
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

	airedEpCount = result.Response.Episodes.AiredCount
	totalEpCount = result.Response.Episodes.TotalCount
	statusValue := result.Response.Status.Value

	// Если статус аниме "вышел" (statusValue == 0), то сайт говорит, что вышло 0 эпизодов.
	// В противном случае - сайт возвращает правильное число.
	if statusValue == 0 {
		return totalEpCount, totalEpCount, nil
	}

	if totalEpCount < airedEpCount {
		totalEpCount = -1
	}

	return airedEpCount, totalEpCount, nil
}

func ParseEpisodes(r io.Reader) (map[int]models.Episode, error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result epJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, err
	}

	eps := make(map[int]models.Episode)
	for _, res := range result.Response {
		epNumber, err := strconv.Atoi(res.Number)
		if err != nil {
			continue
		}

		_, exists := eps[epNumber]
		if !exists {
            ep := models.Episode{}
            ep.EmbedLinks = make(models.EmbedLinks)
			eps[epNumber] = ep
		}

		dubName := res.Data.Dubbing
		u, err := url.Parse(res.IframeUrl)
		if err != nil {
			continue
		}
		playerName := u.Hostname()

		_, exists = eps[epNumber].EmbedLinks[dubName]
		if !exists {
			eps[epNumber].EmbedLinks[dubName] = make(map[string]string)
		}

		eps[epNumber].EmbedLinks[dubName][playerName] = res.IframeUrl
	}

	return eps, nil
}
