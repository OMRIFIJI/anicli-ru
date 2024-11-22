package models
// Структура: озвучка -> плеер -> ссылка на embed
type EmbedLink map[string]map[string]string

// Структура: качество видео -> ссылка на видео
type VideoLink map[string]string

type Episode struct {
	Id        int
	EmbedLink EmbedLink
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
