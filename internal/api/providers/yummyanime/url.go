package yummyanime

import (
    "strconv"
    "net/url"
)

type urlBuilder struct {
	base       string
	searchSuf  string
	searchConf string
	animeSuf   string
	playerSuf  string
	epSuf      string
}

func newUrlBuilder() urlBuilder {
	u := urlBuilder{
		base:       "https://yummy-anime.ru/",
		searchSuf:  "api/search?q=",
		searchConf: "&limit=10&offset=0",
		animeSuf:   "api/anime/",
		playerSuf:  "player?_allow=true",
		epSuf:      "series?id=",
	}
	return u
}

func (u *urlBuilder) searchByTitle(title string) string {
	return u.base + u.searchSuf + url.QueryEscape(title) + u.searchConf
}

func (u *urlBuilder) animeById(id int) string {
	return u.base + u.animeSuf + strconv.Itoa(id)
}

func (u *urlBuilder) animeByUname(uname string) string {
	return u.base + u.animeSuf + url.QueryEscape(uname)
}

func (u *urlBuilder) epById(id string) string {
	return u.base + u.animeSuf + u.epSuf + id
}
