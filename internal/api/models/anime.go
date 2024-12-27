package models

import (
	"errors"
	"sort"
)

// Структура для хранения embed на видео: озвучка -> плеер -> ссылка на embed.
type EmbedLinks map[string]map[string]string

// Структура для хранения прямых ссылок на видео: озвучка -> качество видео -> ссылка на видео.
type VideoLinks map[string]map[int]Video

// Ссылка на видео и опции для его запуска через Mpv.
type Video struct {
	Link    string
	MpvOpts []string
}

// Информация об эпизодах структуры Anime.
type EpisodesContext struct {
	Eps           map[int]*Episode
	EpsSortedKeys []int
	Cur           int
	TotalEpCount  int
	AiredEpCount  int
}

type Episode struct {
	Id         int
	EmbedLinks EmbedLinks
}

// Структура аниме, возвращаемая api.
type Anime struct {
	Id        int
	Uname     string
	Title     string
	MediaType string
	Provider  string
	EpCtx     EpisodesContext
}

// Возвращает ключ выбранного эпизода в Eps (Нумерация начинается с 1).
func (e *EpisodesContext) GetSelectedEpKey() int {
	cur := e.Cur
	key := e.EpsSortedKeys[cur]
	return key
}

func (e *EpisodesContext) GetSelectedEp() (*Episode, error) {
	key := e.GetSelectedEpKey()
	ep, exists := e.Eps[key]
	if !exists {
		return nil, errors.New("Выбранный эпизод не существует")
	}
	return ep, nil
}

func (e *EpisodesContext) SetCur(cur int) error {
	if cur < 0 || cur >= len(e.EpsSortedKeys) {
		return errors.New("Неверное значение курсора")
	}
	e.Cur = cur
	return nil
}

func (e *EpisodesContext) SelectNextEp() error {
	if e.Cur+1 >= len(e.EpsSortedKeys) {
		return errors.New("Вы посмотрели все серии.")
	}
	e.Cur++
	return nil
}

func (e *EpisodesContext) SelectPreviousEp() error {
	if e.Cur-1 < 0 {
		return errors.New("Неверное значение курсора")
	}
	e.Cur--
	return nil
}

func (e *EpisodesContext) SortEpisodeKeys() {
	var episodeKeys []int

	keys := make([]int, 0, len(e.Eps))
	for k := range e.Eps {
		keys = append(keys, k)
	}
	sort.Sort(sort.IntSlice(keys))

	for _, k := range keys {
		episodeKeys = append(episodeKeys, k)
	}

	e.EpsSortedKeys = episodeKeys
}
