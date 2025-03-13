//go:generate stringer -type=PlayerOrigin

package common

import (
	"anicliru/internal/api/models"
	"strings"
)

const DefaultReferer = "https://animego.org"

const (
	AksorDomain   = "aksor.yani.tv"
	AllohaDomain  = "alloha.yani.tv"
	AniboomDomain = "aniboom.one"
	KodikDomain   = "kodik.info"
	SibnetDomain  = "video.sibnet.ru"
	SovromDomain  = "sovetromantica.com"
	VKDomain      = "vk.com"
)

// ENUM
type PlayerOrigin uint

const (
	Aksor PlayerOrigin = iota
	Alloha
	Aniboom
	Kodik
	Sibnet
	Sovrom
	VK
)

const playerOriginName = "aksor.yani.tvalloha.yani.tvaniboom.one"

func NewPlayerOriginMap() map[string]PlayerOrigin {
	return map[string]PlayerOrigin{
		AksorDomain:   Aksor,
		AllohaDomain:  Alloha,
		AniboomDomain: Aniboom,
		KodikDomain:   Kodik,
		SibnetDomain:  Sibnet,
		SovromDomain:  Sovrom,
		VKDomain:      VK,
	}
}

func NewPlayerDomainMap() map[PlayerOrigin]string {
	return map[PlayerOrigin]string{
		Aksor:   AksorDomain,
		Alloha:  AllohaDomain,
		Aniboom: AniboomDomain,
		Kodik:   KodikDomain,
		Sibnet:  SibnetDomain,
		Sovrom:  SovromDomain,
		VK:      VKDomain,
	}
}

type DecodedEmbed struct {
	Video  models.Video
	Origin PlayerOrigin
}

func AppendHttp(url string) string {
	if !strings.HasPrefix(url, "https") {
		return "https:" + url
	}
	return url
}
