package parser

import (
	"anicliru/internal/api/types"
	"golang.org/x/net/html"
	"io"
	"strings"
)

func isAnimeHref(href string) bool {
	return strings.HasPrefix(href, "https://animego.org/anime")
}

func parseAnime(n *html.Node) *types.Anime {
	var href, title string
	for _, attr := range n.Attr {
		if attr.Key == "href" && isAnimeHref(attr.Val) {
			href = attr.Val
		}
		if attr.Key == "title" {
			title = attr.Val
		}
	}

	idInd := strings.LastIndex(href, "-") + 1
	if idInd <= 0 || idInd >= len(href) {
		return nil
	}

	unameInd := strings.LastIndex(href, "/") + 1
	if len(title) > 0 {
		id := href[idInd:]
        uname := href[unameInd:]
		return &types.Anime{
			Id:    id,
            Uname: uname,
			Title: title,
		}
	} else {
		return nil
	}
}

func ParseAnimes(r io.Reader) ([]types.Anime, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var animeSlice []types.Anime
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "a" {
			anime := parseAnime(n)
			if anime != nil {
				animeSlice = append(animeSlice, *anime)
			}
		}
	}

	return animeSlice, nil
}
