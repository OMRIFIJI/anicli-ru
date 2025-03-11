// Было бы неплохо вытягивать не только 1080
package sovrom

import (
	"anicliru/internal/api/models"
	"anicliru/internal/api/player/common"
	httpcommon "anicliru/internal/http"
	"errors"
	"io"
	"regexp"
)

const (
	Netloc = "sovetromantica.com"
)

type Sovrom struct {
	client *httpcommon.HttpClient
}

func NewSovrom() *Sovrom {
	client := httpcommon.NewHttpClient(
		map[string]string{
			"Referer":         common.DefaultReferer,
			"Accept-Language": "ru-RU",
		},
		httpcommon.WithRetries(2),
	)

	a := Sovrom{
		client: client,
	}
	return &a
}

func (a *Sovrom) GetVideos(embedLink string) (map[int]common.DecodedEmbed, error) {
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
	re := regexp.MustCompile(`var\s+config\s*=\s*{\s*\n*\s*"id"\s*:\s*"sovetromantica_player",\s*\n*\s*"file":\s*"(.+?)"`)
	match := re.FindStringSubmatch(string(resBody))

	if match == nil {
		return nil, errors.New("не удалось обработать ссылку на видео")
	}

	link := match[1]

	// Получаю только 1080
	quality := 1080

	links := make(map[int]common.DecodedEmbed)

	links[quality] = common.DecodedEmbed{
		Video:  models.Video{Link: link},
		Origin: Netloc,
	}

	return links, nil
}
