package parser

import (
	"anicliru/internal/api/models"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
)

func parseIdToDub(content string) (map[int]string, error) {
	reEp := regexp.MustCompile(`data-dubbing="(\d+)"[^>]*>\s*<span[^>]*class="video-player-toggle-item-name[^>]*>\s*([^<]+)`)
	matches := reEp.FindAllStringSubmatch(content, -1)

	idToDub := make(map[int]string)
	for _, match := range matches {
		idStr := strings.TrimSpace(match[1])
		id, err := strconv.Atoi(idStr)

		if err != nil {
			continue
		}
		idToDub[id] = strings.TrimSpace(match[2])
	}

	if len(idToDub) == 0 {
		return nil, errors.New("Опции озвучки эпизода не найдены")
	}

	return idToDub, nil
}

// uid озвучки -> плеер -> ссылка
func parseIdToLinks(content string) (map[int]map[string]string, error) {
	reEp := regexp.MustCompile(`data-player="([^\"]*)"\s*data-provider="\d+"\s*data-provide-dubbing="(\d+)"`)
	matches := reEp.FindAllStringSubmatch(content, -1)

	idToLinks := make(map[int]map[string]string)
	for _, match := range matches {
		idStr := strings.TrimSpace(match[2])
		id, err := strconv.Atoi(idStr)

		if err != nil {
			continue
		}

        _, exists := idToLinks[id]
        if !exists{
            idToLinks[id] = make(map[string]string)
        }

        dubLink := strings.TrimSpace(match[1])
		playerName := getPlayerName(dubLink)
        idToLinks[id][playerName] = strings.TrimSpace(match[1])
	}

	if len(idToLinks) == 0 {
		return nil, errors.New("Ссылки на плеер не найдены")
	}

	return idToLinks, nil
}

func ParsePlayerLinks(r io.Reader) (models.EmbedLink, error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result AnimeGoJson
	if err := json.Unmarshal(in, &result); err != nil {
		return nil, err
	}

	idToDub, err := parseIdToDub(result.Content)
	if err != nil {
		return nil, err
	}

	idToLinks, err := parseIdToLinks(result.Content)
	if err != nil {
		return nil, err
	}

	playerLinks := make(models.EmbedLink)
	for id, dubName := range idToDub {
		dubLinks, exists := idToLinks[id]
		if !exists {
			continue
		}
        playerLinks[dubName] = dubLinks
	}

	return playerLinks, nil
}

func getPlayerName(dubLink string) string {
	playerName, exists := strings.CutPrefix(dubLink, "//")
	if !exists {
		return "Неизвестен"
	}
	endOfPlayerName := strings.Index(playerName, ".")
	if endOfPlayerName == -1 {
		return "Неизвестен"
	}

	return playerName[:endOfPlayerName]
}
