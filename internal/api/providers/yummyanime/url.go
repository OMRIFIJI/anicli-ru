package yummyanime

import (
	"net/url"
	"strconv"
)

type urlBuilder struct {
	base       string
	searchSuf  string
	searchConf string
	animeSuf   string
	videoSuf   string
}

func newUrlBuilder() urlBuilder {
	u := urlBuilder{
		base:       "https://yummy-anime.ru/",
		searchSuf:  "api/search?q=",
		searchConf: "&limit=10&offset=0",
		animeSuf:   "api/anime/",
		videoSuf:   "videos",
	}
	return u
}

func (u *urlBuilder) searchByTitle(title string) string {
	return u.base + u.searchSuf + url.QueryEscape(title) + u.searchConf
}

func (u *urlBuilder) animeById(id int) string {
	return u.base + u.animeSuf + strconv.Itoa(id)
}

func (u *urlBuilder) embedByAnimeId(id int) string {
	return u.base + u.animeSuf + strconv.Itoa(id) + "/" + u.videoSuf
}
