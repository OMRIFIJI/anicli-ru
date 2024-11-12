package types

type Episode struct {
	Id int
	// Структура: озвучка -> плеер -> качество видео -> ссылка на видео
	Link map[string]map[string]map[string]*string
}

type Anime struct {
	Id           string
	Uname        string
	Title        string
	Episodes     map[int]Episode
	TotalEpCount int
}

type NotFoundError struct {
	Msg string
}

type HttpError struct {
	Msg string
}

type AnimeError struct {
    Msg string
}

func (e *NotFoundError) Error() string { return e.Msg }
func (e *HttpError) Error() string     { return e.Msg }
func (e *AnimeError) Error() string     { return e.Msg }
