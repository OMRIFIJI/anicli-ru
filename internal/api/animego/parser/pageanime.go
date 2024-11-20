package parser

import (
	"anicliru/internal/api/types"
	"encoding/json"
	"errors"
	"golang.org/x/net/html"
	"io"
	"regexp"
	"strconv"
	"strings"
)

func ParseSeriesEpisodes(r io.Reader) (map[int]int, int, error) {
	episodeIdMap := make(map[int]int)
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
		err := &types.RegionBlockError{
			Msg: "Не доступно на территории РФ",
		}
		return nil, 0, err
	}

	reEp := regexp.MustCompile(`data-episode="(\d+)" *\n* *data-id="(\d+)"`)
	matches := reEp.FindAllStringSubmatch(result.Content, -1)

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

func ParseEpisodeCount(r io.Reader) (epCount int, err error) {
	doc, err := html.Parse(r)
	if err != nil {
		return 0, err
	}

	occurrences := findElements(doc, "dd", "col-6 col-sm-8 mb-1")
	if len(occurrences) == 0 {
		return 0, errors.New("Не удалось найти тег с эпизодами")
	}

	firstDD := occurrences[0]
	for c := firstDD.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode && c.Data == "Фильм" {
			return 1, nil
		}
	}

	if len(occurrences) < 2 {
		return 0, errors.New("Не удалось найти тег с эпизодами")
	}

	secondDD := occurrences[1]
	episodeCount, err := extractNumber(secondDD)
	return episodeCount, err
}

func extractNumber(n *html.Node) (int, error) {
	var textNodeContent string
	var spanContent string

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode && c.Parent.Data != "span" {
			textNodeContent = strings.TrimSpace(c.Data)
		}
		if c.Type == html.ElementNode && c.Data == "span" && c.FirstChild != nil {
			spanContent = strings.TrimSpace(c.FirstChild.Data)
		}
	}

	numberSpan, errSpan := strconv.Atoi(spanContent)
	numberText, errText := strconv.Atoi(textNodeContent)
	// Если онгоинг, то число лежит в span
	if errSpan == nil {
		return numberSpan, nil
	}
	// Если аниме полностью вышло, то просто текст
	if errText == nil {
		return numberText, nil
	}
	return 0, nil
}

func findElements(n *html.Node, tag, class string) []*html.Node {
	var result []*html.Node
	if n.Type == html.ElementNode && n.Data == tag {
		for _, attr := range n.Attr {
			if attr.Key == "class" && attr.Val == class {
				result = append(result, n)
				break
			}
		}
	}
	// Recursively search child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, findElements(c, tag, class)...)
	}
	return result
}
