package anilib

import (
	"errors"
	"strconv"
	"strings"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/common"
)

const Origin = common.Anilib
const headerFields = `--http-header-fields="Referer: https://anilib.me"`

type Anilib struct {
}

func NewAnilib() *Anilib {
	return &Anilib{}
}

// Уже расшифровано
func (a *Anilib) GetVideos(embedLink string) (map[int]common.DecodedEmbed, error) {
	link := common.AppendHttp(embedLink)

	indStart := strings.LastIndex(link, "_")
	if indStart == -1 {
		return nil, errors.New("не удалось найти качество видео")
	}

	indEnd := strings.Index(link, ".mp4")
	if indEnd == -1 {
		return nil, errors.New("не удалось найти качество видео")
	}

	quality, err := strconv.Atoi(link[indStart+1 : indEnd])
	if err != nil {
		return nil, err
	}

	mpvOpts := []string{
		headerFields,
	}

	links := make(map[int]common.DecodedEmbed)
	video := models.Video{
		Link:    link,
		MpvOpts: mpvOpts,
	}
	links[quality] = common.DecodedEmbed{
		Video:  video,
		Origin: Origin,
	}

	return links, nil
}
