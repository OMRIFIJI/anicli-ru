package htmlparser

import (
	"anicliru/internal/types"
	"golang.org/x/net/html"
	"io"
	"strings"
)

func isAnimeHref(href string) bool {
	return strings.HasPrefix(href, "https://animego.org/anime")
}

func getAnime(n *html.Node) *types.Anime {
	var href, title string
	for _, attr := range n.Attr {
		if attr.Key == "href" && isAnimeHref(attr.Val) {
			href = attr.Val
		}
		if attr.Key == "title" {
			title = attr.Val
		}
	}
	if len(href) > 0 && len(title) > 0 {
		return &types.Anime{
			Link:  href,
			Title: title,
		}
	} else {
		return nil
	}
}

func GetAnimes(r io.Reader) ([]types.Anime, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var animeSlice []types.Anime
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "a" {
			anime := getAnime(n)
			if anime != nil {
				animeSlice = append(animeSlice, *anime)
			}
		}
	}

    return animeSlice[1:], nil
}
