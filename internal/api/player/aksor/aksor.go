package aksor

import (
	"anicliru/internal/api/models"
	"anicliru/internal/api/player/common"
	httpcommon "anicliru/internal/http"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const Origin = common.Aksor

type Aksor struct {
	client *httpcommon.HttpClient
}

func NewAksor() *Aksor {
	client := httpcommon.NewHttpClient(
		map[string]string{
			"Referer":         common.DefaultReferer,
			"Accept-Language": "ru-RU",
		},
		httpcommon.WithRetries(2),
	)

	a := Aksor{
		client: client,
	}
	return &a
}

func (a *Aksor) GetVideos(embedLink string) (map[int]common.DecodedEmbed, error) {
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
	re := regexp.MustCompile(`var\s+videoUrl\s*=\s*"(.+?)"`)
	match := re.FindStringSubmatch(string(resBody))

	if match == nil {
		return nil, errors.New("не удалось обработать ссылку на видео")
	}

	link := match[1]

	// Вытягиваю качество видео из ссылки
	qualityStart := strings.LastIndex(link, "/")
	if qualityStart == -1 {
		return nil, errors.New("не удалось обработать ссылку на видео")
	}
	qualityStr := link[qualityStart+1 : len(link)-4]
	quality, err := strconv.Atoi(qualityStr)
	if err != nil {
		return nil, errors.New("не удалось обработать качество видео")
	}

	links := make(map[int]common.DecodedEmbed)

	links[quality] = common.DecodedEmbed{
		Video:  models.Video{Link: link},
		Origin: Origin,
	}

	return links, nil
}
