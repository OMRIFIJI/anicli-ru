package sibnet

import (
	"anicliru/internal/api/models"
	"anicliru/internal/api/player/common"
	httpcommon "anicliru/internal/http"
	"errors"
	"fmt"
	"io"
	"regexp"
)

const Netloc = "video.sibnet.ru"

type Sibnet struct {
	client  *httpcommon.HttpClient
	baseUrl string
}

func NewSibnet() *Sibnet {
	client := httpcommon.NewHttpClient(
		map[string]string{
			"Referer":                   "https://animego.org/",
			"User-Agent":                "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0",
			"Upgrade-Insecure-Requests": "1",
			"Accept-Language":           "ru-RU",
		},
		httpcommon.WithRetries(5), // Sibnet - любитель помолчать
		httpcommon.WithTimeout(2), // Очень редко может не отвечать больше 3 раз, но ждать его не очень хочется...
	)
	s := &Sibnet{
		client:  client,
		baseUrl: "https://video.sibnet.ru/",
	}
	return s
}

func (s *Sibnet) GetVideos(embedLink string) (map[int]common.DecodedEmbed, error) {
	embedLink = "https:" + embedLink
	res, err := s.client.Get(embedLink)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`src:\s*"\/(.+?\.mp4)"`)
	match := re.FindStringSubmatch(string(resBody))

	if match == nil {
		return nil, errors.New("Не удалось обработать ссылку на видео")
	}
	link := s.baseUrl + match[1]

	links := make(map[int]common.DecodedEmbed)
	headersOpt := fmt.Sprintf(`--http-header-fields="Referer: %s","Upgrade-Insecure-Requests: 1"`, embedLink)

    // Всего одна ссылка
    video := models.Video{
		Link:    link,
		MpvOpts: []string{headersOpt},
	}
	links[480] = common.DecodedEmbed{
        Video: video,
        Origin: Netloc,
    }
	return links, nil
}
