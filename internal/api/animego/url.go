package animego

type urlBuilder struct {
	base       string
	searchSuf  string
	animeSuf   string
	playerSuf  string
	episodeSuf string
}

func newUrlBuilder() *urlBuilder {
	u := urlBuilder{
		base:       "https://animego.org/",
		searchSuf:  "search/anime?q=",
		animeSuf:   "anime/",
		playerSuf:  "player?_allow=true",
		episodeSuf: "series?id=",
	}
	return &u
}

func (u *urlBuilder) searchByTitle(title string) string {
	return u.base + u.searchSuf + title
}

func (u *urlBuilder) animeById(id string) string {
	return u.base + u.animeSuf + id + "/" + u.playerSuf
}

func (u *urlBuilder) animeByUname(uname string) string {
	return u.base + u.animeSuf + uname
}

func (u *urlBuilder) episodeById(id string) string {
	return u.base + u.animeSuf + u.episodeSuf + id
}
