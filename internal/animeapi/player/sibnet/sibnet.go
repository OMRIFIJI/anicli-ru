package sibnet

import (
	"errors"
	"fmt"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/common"
	httpkit "github.com/OMRIFIJI/anicli-ru/internal/httpkit"
	"io"
	"regexp"
)

const (
	Origin  = common.Sibnet
	baseUrl = "https://video.sibnet.ru"
)

type Sibnet struct {
	client  *httpkit.HttpClient
	baseUrl string
}

func NewSibnet() *Sibnet {
	client := httpkit.NewHttpClient(
		map[string]string{
			"Referer":                   common.DefaultReferer,
			"User-Agent":                "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0",
			"Upgrade-Insecure-Requests": "1",
			"Accept-Language":           "ru-RU",
		},
		httpkit.WithRetries(5), // Sibnet - любитель помолчать
		httpkit.WithTimeout(2), // Очень редко может не отвечать больше 3 раз, но ждать его не очень хочется...
	)
	s := &Sibnet{
		client:  client,
		baseUrl: baseUrl,
	}
	return s
}

func (s *Sibnet) GetVideos(embedLink string) (map[int]common.DecodedEmbed, error) {
	embedLink = common.AppendHttp(embedLink)

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
		return nil, errors.New("не удалось обработать ссылку на видео")
	}
	link := s.baseUrl + match[1]

	links := make(map[int]common.DecodedEmbed)
	headerFields := fmt.Sprintf(`--http-header-fields="Referer: %s","Upgrade-Insecure-Requests: 1"`, embedLink)

	// Всего одна ссылка
	video := models.Video{
		Link:    link,
		MpvOpts: []string{headerFields},
	}
	links[480] = common.DecodedEmbed{
		Video:  video,
		Origin: Origin,
	}
	return links, nil
}
