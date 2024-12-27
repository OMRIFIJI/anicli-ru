package animego

import (
	"net/url"
	"strconv"
)

type urlBuilder struct {
	base      string
	searchSuf string
	animeSuf  string
	playerSuf string
	epSuf     string
}

func newUrlBuilder() urlBuilder {
	u := urlBuilder{
		base:      "https://animego.org/",
		searchSuf: "search/anime?q=",
		animeSuf:  "anime/",
		playerSuf: "player?_allow=true",
		epSuf:     "series?id=",
	}
	return u
}

func (u *urlBuilder) searchByTitle(title string) string {
	return u.base + u.searchSuf + url.QueryEscape(title)
}

func (u *urlBuilder) animeById(id int) string {
	return u.base + u.animeSuf + strconv.Itoa(id) + "/" + u.playerSuf
}

func (u *urlBuilder) animeByUname(uname string) string {
	return u.base + u.animeSuf + url.QueryEscape(uname)
}

func (u *urlBuilder) epById(id string) string {
	return u.base + u.animeSuf + u.epSuf + id
}
