package vk

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

const Origin = common.VK

type VK struct {
	client *httpcommon.HttpClient
}

func NewVK() *VK {
	client := httpcommon.NewHttpClient(
		map[string]string{
			"Accept-Language": "ru-RU",
		},
		httpcommon.WithRetries(2),
	)
	vk := &VK{
		client: client,
	}
	return vk
}

func (vk *VK) GetVideos(embedLink string) (map[int]common.DecodedEmbed, error) {
	embedLink = common.AppendHttp(embedLink)

	res, err := vk.client.Get(embedLink)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`"url(\d+)"\s*:\s*"(.+?)"`)
	matches := re.FindAllStringSubmatch(string(resBody), -1)

	if len(matches) == 0 {
		return nil, errors.New("не удалось обработать ссылки на видео")
	}

	links := make(map[int]common.DecodedEmbed)
	for _, match := range matches {
		quality, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, errors.New("ошибка обработки качества видео")
		}
		link := removeSlashes(match[2])

		video := models.Video{
			Link:    link,
			MpvOpts: nil,
		}
		links[quality] = common.DecodedEmbed{
			Video:  video,
			Origin: Origin,
		}
	}
	return links, nil
}

func removeSlashes(link string) string {
	link = strings.ReplaceAll(link, `\`, "")
	return link
}
