package common

import (
	"strings"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
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
	AnilibDomain  = "video1.anilib.me"
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
	Anilib
)

func GetPlayerDomains() []string {
	return []string{
		AksorDomain,
		AllohaDomain,
		AniboomDomain,
		KodikDomain,
		SibnetDomain,
		SovromDomain,
		VKDomain,
		AnilibDomain,
	}
}

func NewPlayerOriginMap() map[string]PlayerOrigin {
	return map[string]PlayerOrigin{
		AksorDomain:   Aksor,
		AllohaDomain:  Alloha,
		AniboomDomain: Aniboom,
		KodikDomain:   Kodik,
		SibnetDomain:  Sibnet,
		SovromDomain:  Sovrom,
		VKDomain:      VK,
		AnilibDomain:  Anilib,
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
		Anilib:  AnilibDomain,
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
