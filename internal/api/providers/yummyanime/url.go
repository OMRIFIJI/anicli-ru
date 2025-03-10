package yummyanime

import (
	"fmt"
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

func newUrlBuilder(fullDomain string) urlBuilder {
	u := urlBuilder{
		base:       fmt.Sprintf("https://%s/", fullDomain),
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
