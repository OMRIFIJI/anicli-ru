package parser

import (
	"github.com/OMRIFIJI/anicli-ru/internal/api/models"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func parseAnime(n *html.Node, searchPos int, animePageUrl string) *models.Anime {
	var href, title string
	for _, attr := range n.Attr {
		if attr.Key == "href" && strings.HasPrefix(attr.Val, animePageUrl) {
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
			Id:        id,
			Uname:     uname,
			Title:     title,
			Provider:  "animego",
			SearchPos: searchPos,
		}
	} else {
		return nil
	}
}

func ParseAnimes(r io.Reader, urlBase string) ([]*models.Anime, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var animeSlice []*models.Anime
	animePageUrl := fmt.Sprintf("%sanime", urlBase)
	searchPos := 0
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "a" {
			anime := parseAnime(n, searchPos, animePageUrl)
			if anime != nil {
				animeSlice = append(animeSlice, anime)
				searchPos++
			}
		}
	}

	return animeSlice, nil
}
