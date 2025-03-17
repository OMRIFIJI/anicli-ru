package parser

// Облегченная структура json'a, возвращаемого в результате поиска аниме.
type searchJson struct {
	Data []FoundAnime `json:"data"`
}

type FoundAnime struct {
	Title string    `json:"rus_name"`
	Id    int       `json:"id"`
	Slug  string    `json:"slug"`
	Type  animeType `json:"type"`
}

type animeType struct {
	Label string `json:"label"`
}

// Облегченная структура json'a,
// возвращаемого в результате запроса информации об аниме по его id.
type animeJson struct {
	Data animeInfo `json:"data"`
}

type animeInfo struct {
	ItemsCount epCountInfo `json:"items_count"`
}

type epCountInfo struct {
	TotalCount    int `json:"total"`
	UploadedCount int `json:"uploaded"`
}

// Структура для получения информации об id эпизодов
type epMetadataJson struct {
	Data []epMetadata `json:"data"`
}

type epMetadata struct {
	Id     int    `json:"id"`
	Number int `json:"item_number"`
}

type epVideoJson struct {
	Data epVideoData `json:"data"`
}

type epVideoData struct {
	Players []epVideo `json:"players"`
}

// Страшный композит для получения embed'ов
// В случае плеера animelib embed в Video,
// в противном случае в Src
type epVideo struct {
	TsType translationType `json:"translation_type"`
	Team   dubTeam         `json:"team"`
	Src    string          `json:"src"`
	Video  *videoModel     `json:"video,omitempty"`
}

type translationType struct {
	Label string `json:"label"`
}

type dubTeam struct {
	Name string `json:"name"`
}

type videoModel struct {
	Quality []videoWithQuality `json:"quality"`
}

type videoWithQuality struct {
	Quality int    `json:"quality"`
	Href    string `json:"href"`
}
