package types

type EpisodeLinks map[EpisodeQuality]EpisodeLink
type EpisodeQuality string
type PlayerName string
type EpisodeLink string

type Anime struct {
	Link     string
	Title    string
	Episodes []Episode
}

type Episode interface {
    GetLink(EpisodeQuality, PlayerName) string
}

type Parser interface {
	FindAnimeByTitle(title string) ([]Anime, error)
}
