package parser

import (
	"anicliru/internal/api/models"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func isAnimeHref(href string) bool {
	return strings.HasPrefix(href, "https://animego.org/anime")
}

func parseAnime(n *html.Node) *models.Anime {
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
        id, err := strconv.Atoi(href[idInd:])
        if err != nil {
            return nil
        }
		uname := href[unameInd:]
		return &models.Anime{
			Id:    id,
			Uname: uname,
			Title: title,
		}
	} else {
		return nil
	}
}

func ParseAnimes(r io.Reader) ([]*models.Anime, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var animeSlice []*models.Anime
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "a" {
			anime := parseAnime(n)
			if anime != nil {
				animeSlice = append(animeSlice, anime)
			}
		}
	}

	if len(animeSlice) == 0 {
		notFoundError := models.NotFoundError{
			Msg: "По вашему запросу не удалось ничего найти.",
		}
		return nil, &notFoundError
	}

	return animeSlice, nil
}
