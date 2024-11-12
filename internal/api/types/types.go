package types

type Episode struct {
	Id int
	// Структура: озвучка -> плеер -> качество видео -> ссылка на видео
	Link map[string]map[string]map[string]*string
}

type Anime struct {
	Id       string
	Title    string
	Episodes map[int]Episode
}

type NotFoundError struct {
	Msg string
}

func (e *NotFoundError) Error() string { return e.Msg }

type NoConnectionError struct {
	Msg string
}

func (e *NoConnectionError) Error() string { return e.Msg }
