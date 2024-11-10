package types

type Anime struct {
	Link     string
	Title    string
	Episodes []Episode
}

type Episode struct {
	videoLinks EpisodeLinks
}

type Parser interface {
	FindAnimeByTitle(title string) ([]Anime, error)
}

type EpisodeLinks map[EpisodeQuality]EpisodeLink
type EpisodeQuality string
type EpisodeLink string
