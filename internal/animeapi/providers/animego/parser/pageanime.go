package parser

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
)

func ParseEpIds(r io.Reader) (epIdMap map[int]int, err error) {
	epIdMap = make(map[int]int)
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result animeGoJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, err
	}

	re := regexp.MustCompile("Видео недоступно на территории")
	match := re.FindString(result.Content)
	if len(match) != 0 {
		return nil, errors.New("не доступно на территории РФ")
	}

	re = regexp.MustCompile(`data-episode="(\d+)"\s*\n*\s*data-id="(\d+)"`)
	matches := re.FindAllStringSubmatch(result.Content, -1)

	for _, match := range matches {
		epNum, errNum := strconv.Atoi(match[1])
		epId, errId := strconv.Atoi(match[2])

		if errNum != nil || errId != nil {
			continue
		}

		epIdMap[epNum] = epId
	}

	if len(epIdMap) == 0 {
		return nil, errors.New("нет информации ни об одной серии")

	}

	return epIdMap, nil
}

func ParseFilmRegionBlock(r io.Reader) (isRegionBlocked bool, err error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return false, err
	}

	var result animeGoJson
	if err := json.Unmarshal(in, &result); err != nil {
		return false, err
	}

	re := regexp.MustCompile("Видео недоступно на территории")
	match := re.FindString(result.Content)
	if len(match) != 0 {
		return true, nil
	}
	return false, nil
}

func IsValidEp(r io.Reader) bool {
	in, err := io.ReadAll(r)
	if err != nil {
		return false
	}

	var result animeGoJson
	if err := json.Unmarshal(in, &result); err != nil {
		return false
	}

	trimmed := strings.Trim(result.Content, " \t\n")
	return trimmed != ""
}

func ParseMediaStatus(r io.Reader) (AiredEpCount, TotalEpCount int, mediaType string, err error) {
	resultByte, err := io.ReadAll(r)
	if err != nil {
		return -1, -1, "", err
	}
	result := string(resultByte)

	re := regexp.MustCompile(`Тип\s*<\/dt>\s*\n*\s*<dd.+?>(.+?)<`)
	match := re.FindStringSubmatch(result)
	if match == nil {
		return -1, -1, "", nil
	}

	mediaType = strings.TrimSpace(match[1])
	mediaType = strings.ToLower(mediaType)

	// Если онгоинг группы 1 и 4 - текущее и общее количество,
	// Если вышло, то в группе 1 - общее количество
	re = regexp.MustCompile(`Эпизоды\s*<\/dt>\s*\n*\s*<dd.+?">(\d*)\s*(.*?|\/)\s*(<span>|.*?)\s*(\d*)\s*(<\/span>|<\/dd>)`)
	match = re.FindStringSubmatch(result)
	if match == nil {
		return -1, -1, mediaType, nil
	}

	AiredEpCount, err = strconv.Atoi(match[1])
	if err != nil {
		return -1, -1, mediaType, nil
	}

	// Если онгоинг
	if len(match[4]) != 0 {
		TotalEpCount, err = strconv.Atoi(match[4])
		if err != nil {
			return AiredEpCount, -1, mediaType, nil
		}
		return AiredEpCount, TotalEpCount, mediaType, nil
	}

	return AiredEpCount, AiredEpCount, mediaType, nil
}
