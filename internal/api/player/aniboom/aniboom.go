package aniboom

import (
	"github.com/OMRIFIJI/anicli-ru/internal/api/models"
	"github.com/OMRIFIJI/anicli-ru/internal/api/player/common"
	httpkit "github.com/OMRIFIJI/anicli-ru/internal/httpkit"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const (
	Origin       = common.Aniboom
	baseUrl      = "https://aniboom.one"
	headerFields = `--http-header-fields="Referer: https://aniboom.one","Accept-Language: ru-RU"`
)

type Aniboom struct {
	client     *httpkit.HttpClient
	clientDash *httpkit.HttpClient
}

func NewAniboom() *Aniboom {
	client := httpkit.NewHttpClient(
		map[string]string{
			"Referer":         common.DefaultReferer,
			"Accept-Language": "ru-RU",
		},
		httpkit.WithRetries(2),
	)
	clientDash := httpkit.NewHttpClient(
		map[string]string{
			"Referer":         baseUrl,
			"Accept-Language": "ru-RU",
		},
	)

	a := Aniboom{
		client:     client,
		clientDash: clientDash,
	}
	return &a
}

func (a *Aniboom) GetVideos(embedLink string) (map[int]common.DecodedEmbed, error) {
	embedLink = common.AppendHttp(embedLink)
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

	if matchDash == nil {
		return nil, errors.New("не удалось обработать ссылку на видео")
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

	links := make(map[int]common.DecodedEmbed)
	for i, match := range matchOpts {
		quality, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}

		mpvOpts := []string{
			headerFields,
			fmt.Sprintf("--vid=%d", i+1),
		}

		video := models.Video{
			Link:    dash,
			MpvOpts: mpvOpts,
		}
		links[quality] = common.DecodedEmbed{
			Video:  video,
			Origin: Origin,
		}
	}

	if len(links) == 0 {
		return nil, errors.New("не удалось обработать опции видео")
	}

	return links, nil
}

func removeSlashes(link string) string {
	link = strings.ReplaceAll(link, `\\\`, "")
	return link
}
