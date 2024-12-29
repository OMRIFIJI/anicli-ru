package parser

import (
	"anicliru/internal/api/models"
	"encoding/json"
	"io"
	"net/url"
	"strconv"
)

func ParseAnimes(r io.Reader) (*SearchJson, error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result SearchJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, err
	}

	return &result, nil
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

    if (totalEpCount < airedEpCount) {
        totalEpCount = -1
    }

    return airedEpCount, totalEpCount, nil
}

func ParseEpisodes(r io.Reader) (map[int]*models.Episode, error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result EpJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, err
	}

    eps := make(map[int]*models.Episode)
	for _, res := range result.Response {
        epNumber, err := strconv.Atoi(res.Number)
        if err != nil {
            continue
        }

        _, exists := eps[epNumber]
        if !exists {
            eps[epNumber] = &models.Episode{}
            eps[epNumber].EmbedLinks = make(models.EmbedLinks)
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
