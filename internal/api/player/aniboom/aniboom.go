package aniboom

import (
	httpcommon "anicliru/internal/http"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Aniboom struct {
	client *httpcommon.HttpClient
}

func NewAniboom() *Aniboom {
	client := httpcommon.NewHttpClient(
		map[string]string{
			"Referer": "https://animego.org/",
		},
	)

	a := Aniboom{
		client: client,
	}
	return &a
}

func (a *Aniboom) FindLinks(embedLink string) (map[int]string, error) {
	embedLink = "https:" + embedLink
	res, err := a.client.Get(embedLink)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
    if err != nil {
        return nil, err
    }

	// Отложил до лучших времён, не хочу использовать m3u8 с mpv пока не разрешится эта проблема:
	// https://github.com/mpv-player/mpv/issues/11441
	// reHls := regexp.MustCompile(`&quot;hls&quot;:&quot;{\\&quot;src\\&quot;:\\&quot;(https[^;]+?\.m3u8)\\&quot;`)
	// matchHls := reHls.FindStringSubmatch(string(resBody))

	reDash := regexp.MustCompile(`&quot;dash&quot;:&quot;{\\&quot;src\\&quot;:\\&quot;(https[^;]+?\.mpd)\\&quot;`)
	matchDash := reDash.FindStringSubmatch(string(resBody))

	reQuality := regexp.MustCompile(`&quot;qualityVideo&quot;:(\d+)`)
	matchQuality := reQuality.FindStringSubmatch(string(resBody))

	if matchDash != nil && matchQuality != nil {
		dash := removeSlashes(matchDash[1])
		quality, err := strconv.Atoi(matchQuality[1])
		if err != nil {
			return nil, err
		}

		links := make(map[int]string)
		links[quality] = dash
		return links, nil
	}

	return nil, errors.New("Не удалось обработать видео или его качество")
}

func removeSlashes(link string) string {
	link = strings.ReplaceAll(link, `\\\`, "")
	return link
}
