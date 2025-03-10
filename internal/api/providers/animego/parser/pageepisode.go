package parser

import (
	"anicliru/internal/api/models"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
)

func parseIdToDub(content string) (map[string]string, error) {
	reEp := regexp.MustCompile(`data-dubbing="(.+?)"[^>]*>\s*<span[^>]*class="video-player-toggle-item-name[^>]*>\s*([^<]+)`)
	matches := reEp.FindAllStringSubmatch(content, -1)

	idToDub := make(map[string]string)
	for _, match := range matches {
		id := strings.TrimSpace(match[1])
		idToDub[id] = strings.TrimSpace(match[2])
	}

	if len(idToDub) == 0 {
		return nil, errors.New("опции озвучки эпизода не найдены")
	}

	return idToDub, nil
}

// uid озвучки -> плеер -> ссылка
func parseIdToLinks(content string) (map[string]map[string]string, error) {
	reEp := regexp.MustCompile(`data-player="(.+?)"\s*data-provider=".+?"\s*data-provide-dubbing="(.+?)"`)
	matches := reEp.FindAllStringSubmatch(content, -1)

	idToLinks := make(map[string]map[string]string)
	for _, match := range matches {
		id := strings.TrimSpace(match[2])

        _, exists := idToLinks[id]
        if !exists{
            idToLinks[id] = make(map[string]string)
        }

        dubLink := strings.TrimSpace(match[1])
        dubLink = strings.ReplaceAll(dubLink, "&amp;", "&")
		playerName := netloc(dubLink)
        idToLinks[id][playerName] = dubLink
	}

	if len(idToLinks) == 0 {
		return nil, errors.New("Ссылки на плеер не найдены")
	}

	return idToLinks, nil
}

func ParseEmbedLinks(r io.Reader) (models.EmbedLinks, error) {
	in, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result animeGoJson
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

	embedLinks := make(models.EmbedLinks)
	for id, dubName := range idToDub {
		dubLinks, exists := idToLinks[id]
		if !exists {
			continue
		}
        embedLinks[dubName] = dubLinks
	}

	return embedLinks, nil
}

func netloc(dubLink string) string {
	playerName, exists := strings.CutPrefix(dubLink, "//")
	if !exists {
		return "Неизвестен"
	}
	endOfPlayerName := strings.Index(playerName, "/")
	if endOfPlayerName == -1 {
		return "Неизвестен"
	}

	return playerName[:endOfPlayerName]
}
