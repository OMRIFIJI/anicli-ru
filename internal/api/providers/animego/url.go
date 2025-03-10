package animego

import (
	"fmt"
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

func newUrlBuilder(fullDomain string) urlBuilder {
	u := urlBuilder{
		base:      fmt.Sprintf("https://%s/", fullDomain),
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
