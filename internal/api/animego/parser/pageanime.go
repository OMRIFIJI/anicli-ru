package parser

import (
	"anicliru/internal/api/types"
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type AnimeGoJson struct {
	Content string `json:"content"`
}

func ParseEpisodes(r io.Reader) (map[int]int, int, error) {
	episodeIdMap := make(map[int]int)
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, 0, err
	}

	var result AnimeGoJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, 0, err
	}

	re := regexp.MustCompile(`data-episode="(\d+)" *\n* *data-id="(\d+)"`)
	matches := re.FindAllStringSubmatch(result.Content, -1)

	var lastEpisodeNum int
	for _, match := range matches {
		episodeNum, errNum := strconv.Atoi(match[1])
		episodeId, errId := strconv.Atoi(match[2])

		if errNum != nil || errId != nil {
			continue
		}

		episodeIdMap[episodeNum] = episodeId
		lastEpisodeNum = episodeNum
	}

	if len(episodeIdMap) == 0 {
		err := types.NotFoundError{
			Msg: "Нет информации ни об одной серии.",
		}
		return nil, 0, &err
	}

	return episodeIdMap, lastEpisodeNum, nil
}

func IsValid(r io.Reader) bool {

    in, err := io.ReadAll(r)
	if err != nil {
		return false
	}

	var result AnimeGoJson
	if err := json.Unmarshal(in, &result); err != nil {
		return false
	}

    trimmed := strings.Trim(result.Content, " \t\n")
    return trimmed == ""
}
