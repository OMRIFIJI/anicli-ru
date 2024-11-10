package animego

import (
	"anicliru/internal/api/types"
	"net/http"
)

type URL struct {
	base string
	suff URLSuffixes
}

type URLSuffixes struct {
	search string
}

type AnimeGo struct {
	client http.Client
	url    URL
	animes []types.Anime
	title  string
}
