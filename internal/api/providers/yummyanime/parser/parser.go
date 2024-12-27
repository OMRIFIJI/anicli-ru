package parser

import (
	"encoding/json"
	"io"
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

    return airedEpCount, totalEpCount, nil
}
