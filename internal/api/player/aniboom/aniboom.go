package aniboom

import (
	"anicliru/internal/api/models"
	httpcommon "anicliru/internal/http"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Aniboom struct {
	client     *httpcommon.HttpClient
	clientDash *httpcommon.HttpClient
}

func NewAniboom() *Aniboom {
	client := httpcommon.NewHttpClient(
		map[string]string{
			"Referer":         "https://animego.org/",
			"Accept-Language": "ru-RU",
		},
	)
	clientDash := httpcommon.NewHttpClient(
		map[string]string{
			"Referer":         "https://aniboom.one",
			"Accept-Language": "ru-RU",
		},
	)

	a := Aniboom{
		client:     client,
		clientDash: clientDash,
	}
	return &a
}

func (a *Aniboom) FindVideos(embedLinks string) (map[int]models.Video, error) {
	embedLinks = "https:" + embedLinks
	res, err := a.client.Get(embedLinks)
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

	if matchDash == nil {
		return nil, errors.New("Не удалось обработать ссылку на видео")
	}

	dash := removeSlashes(matchDash[1])

	resDash, err := a.clientDash.Get(dash)
	if err != nil {
		return nil, err
	}
	defer resDash.Body.Close()

	resBody, err = io.ReadAll(resDash.Body)
	if err != nil {
		return nil, err
	}

	reOpts := regexp.MustCompile(`mimeType="video.*?".*?\s+width="\d+"\s+height="(\d+)"`)
	matchOpts := reOpts.FindAllStringSubmatch(string(resBody), -1)

	links := make(map[int]models.Video)
    headersOpt := `--http-header-fields="Referer: https://aniboom.one","Accept-Language: ru-RU"`
	for i, match := range matchOpts {
		quality, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}

		mpvOpts := []string{
            headersOpt,
			fmt.Sprintf("--vid=%d", i+1),
		}

		links[quality] = models.Video{
			Link:    dash,
			MpvOpts: mpvOpts,
		}
	}

	if len(links) == 0 {
		return nil, errors.New("Не удалось обработать опции видео")
	}

	return links, nil
}

func removeSlashes(link string) string {
	link = strings.ReplaceAll(link, `\\\`, "")
	return link
}
