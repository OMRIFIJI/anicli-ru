package parser

import (
	"anicliru/internal/api/models"
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"
)

func ParseEpIds(r io.Reader) (epIdMap map[int]int, lastEpNum int, err error) {
	epIdMap = make(map[int]int)
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, 0, err
	}

	var result AnimeGoJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, 0, err
	}

	reCheck := regexp.MustCompile("Видео недоступно на территории")
	match := reCheck.FindString(result.Content)
	if len(match) != 0 {
		err := &models.RegionBlockError{
			Msg: "Не доступно на территории РФ",
		}
		return nil, 0, err
	}

	reEp := regexp.MustCompile(`data-episode="(\d+)"\s*\n*\s*data-id="(\d+)"`)
	matches := reEp.FindAllStringSubmatch(result.Content, -1)

	for _, match := range matches {
		epNum, errNum := strconv.Atoi(match[1])
		epId, errId := strconv.Atoi(match[2])

		if errNum != nil || errId != nil {
			continue
		}

		epIdMap[epNum] = epId
		lastEpNum = epNum
	}

	if len(epIdMap) == 0 {
		err := models.NotFoundError{
			Msg: "Нет информации ни об одной серии.",
		}
		return nil, 0, &err
	}

	return epIdMap, lastEpNum, nil
}

func ParseFilmRegionBlock(r io.Reader) (isRegionBlocked bool, err error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return false, err
	}

	var result AnimeGoJson
	if err := json.Unmarshal(in, &result); err != nil {
		return false, err
	}

	reCheck := regexp.MustCompile("Видео недоступно на территории")
	match := reCheck.FindString(result.Content)
	if len(match) != 0 {
		return true, nil
	}
	return false, nil
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
	return trimmed != ""
}

func ParseMediaStatus(r io.Reader) (epCount int, mediaType string, err error) {
	resultByte, err := io.ReadAll(r)
	if err != nil {
		return -1, "", err
	}
    result := string(resultByte)

    reCheck := regexp.MustCompile(`Тип\s*<\/dt>\s*\n*\s*<dd.+?>(.+?)<`)
	match := reCheck.FindStringSubmatch(result)
	if match == nil {
		return -1, "", nil
	}

    mediaType = strings.TrimSpace(match[1])
    mediaType = strings.ToLower(mediaType)

    reCheck = regexp.MustCompile(`Эпизоды\s*<\/dt>\s*\n*\s*<dd.+?>(\d+)<`)
	match = reCheck.FindStringSubmatch(result)
	if match == nil {
		return -1, mediaType, nil
	}

    // Здесь либо 2 числа, если онгоинг, либо 1, если вышло
    epCount, err = strconv.Atoi(match[1])
    if err != nil {
		return -1, mediaType, nil
    }
	return epCount, mediaType, nil
}
