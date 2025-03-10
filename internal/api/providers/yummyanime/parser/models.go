package parser

// Облегченная структура json'a, возвращаемого в результате поиска аниме.
type SearchJson struct {
	Response []FoundAnime `json:"response"`
}

type FoundAnime struct {
	Title    string    `json:"title"`
	AnimeUrl string    `json:"anime_url"`
	Id       int       `json:"anime_id"`
	Type     animeType `json:"type"`
}

type animeType struct {
	Name string `json:"name"`
}

// Облегченная структура json'a,
// возвращаемого в результате запроса информации об аниме по его id.
type animeJson struct {
	Response animeInfo `json:"response"`
}

type animeInfo struct {
	Status   animeStatus `json:"anime_status"`
	Episodes episodeInfo `json:"episodes"`
}

// Значение 0 соответствует "вышел".
type animeStatus struct {
	Value int `json:"value"`
}

type episodeInfo struct {
	TotalCount int `json:"count"`
	AiredCount int `json:"aired"`
}

// Структура для получения информации об embed на видео
type EpJson struct {
	Response []foundEpisode `json:"response"`
}

type foundEpisode struct {
	Number    string      `json:"number"`
	IframeUrl string      `json:"iframe_url"`
	Data      episodeData `json:"data"`
}

type episodeData struct {
	Dubbing string `json:"dubbing"`
}
