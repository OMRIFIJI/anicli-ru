package anilib

import (
	"fmt"
	"net/url"
	"strconv"
)

type urlBuilder struct {
	base           string
	animeSearchSuf string
	epSearchSuf    string
	animeSuf       string
	animeCfgSuf    string
	epSuf          string
}

func newUrlBuilder(fullDomain string) urlBuilder {
	u := urlBuilder{
		base:           fmt.Sprintf("https://%s/", fullDomain),
		animeSearchSuf: "api/anime?q=",
		epSearchSuf:    "api/episodes?anime_id=",
		animeSuf:       "api/anime/",
		animeCfgSuf:    "?fields[]=episodes_count",
		epSuf:          "api/episodes/",
	}
	return u
}

func (u *urlBuilder) searchByTitle(title string) string {
	return u.base + u.animeSearchSuf + url.QueryEscape(title)
}

func (u *urlBuilder) animeBySlugUrl(slugUrl string) string {
	return u.base + u.animeSuf + slugUrl + u.animeCfgSuf
}

func (u *urlBuilder) epsIdBySlugUrl(slugUrl string) string {
	return u.base + u.epSearchSuf + slugUrl
}

func (u *urlBuilder) embedByEpId(id int) string {
	return u.base + u.epSuf + strconv.Itoa(id)
}
