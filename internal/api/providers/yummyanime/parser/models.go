package parser

// Облегченная структура json'a, возвращаемого в результате поиска аниме.
type SearchJson struct {
	Response []foundAnime `json:"response"`
}

type foundAnime struct {
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
    Status animeStatus `json:"anime_status"`
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
