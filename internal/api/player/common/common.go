package common

import "anicliru/internal/api/models"

const DefaultReferer = "https://animego.org"

type DecodedEmbed struct {
	Video  models.Video
	Origin string
}
