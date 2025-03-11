package common

import (
	"anicliru/internal/api/models"
	"strings"
)

const DefaultReferer = "https://animego.org"

type DecodedEmbed struct {
	Video  models.Video
	Origin string
}

func AppendHttp(url string) string {
	if !strings.HasPrefix(url, "https") {
		return "https:" + url
	}
	return url
}
