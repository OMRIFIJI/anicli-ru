package types

// Структура: озвучка -> плеер -> качество видео -> ссылка на видео
type EpisodeLinks map[string]map[string]map[string]*string

// Структура: озвучка -> плеер -> ссылка на embed
type PlayerLinks map[string]map[string]string

type Episode struct {
	Id         int
	EpLink     EpisodeLinks
	PlayerLink PlayerLinks
}

type Anime struct {
	Id           string
	Uname        string
	Title        string
	Episodes     map[int]*Episode
	TotalEpCount int
}

type NotFoundError struct {
	Msg string
}

type NotAvailableError struct {
	Msg string
}

type HttpError struct {
	Msg string
}

type ParseError struct {
	Msg string
}

type RegionBlockError struct {
	Msg string
}

func (e *NotFoundError) Error() string     { return e.Msg }
func (e *NotAvailableError) Error() string { return e.Msg }
func (e *HttpError) Error() string         { return e.Msg }
func (e *ParseError) Error() string        { return e.Msg }
func (e *RegionBlockError) Error() string  { return e.Msg }
