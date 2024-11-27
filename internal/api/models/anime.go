package models

import (
	"errors"
	"sort"
)

const FilmEpisodeId = -1

// Структура: озвучка -> плеер -> ссылка на embed
type EmbedLinks map[string]map[string]string

// Структура: озвучка -> качество видео -> ссылка на видео
type VideoLinks map[string]map[int]Video

type Video struct {
	Link string
	MpvOpts []string
}

type EpisodesContext struct {
	Eps           map[int]*Episode
	EpsSortedKeys []int
	Cur           int
	TotalEpCount  int
}

type Episode struct {
	Id         int
	EmbedLinks EmbedLinks
}

type Anime struct {
	Id        int
	Uname     string
	Title     string
	MediaType string
	EpCtx     EpisodesContext
}

func (a *Anime) GetSelectedEpKey() int {
	cur := a.EpCtx.Cur
	key := a.EpCtx.EpsSortedKeys[cur]
	return key
}

func (a *Anime) GetSelectedEp() (*Episode, error) {
	key := a.GetSelectedEpKey()
	ep, exists := a.EpCtx.Eps[key]
	if !exists {
		return nil, errors.New("Выбранный эпизод не существует")
	}
	return ep, nil
}

func (a *Anime) SetCur(cur int) error {
	if cur < 0 || cur >= len(a.EpCtx.EpsSortedKeys) {
		return errors.New("Неверное значение курсора")
	}
	a.EpCtx.Cur = cur
	return nil
}

func (a *Anime) SelectNextEp() error {
	if a.EpCtx.Cur+1 >= len(a.EpCtx.EpsSortedKeys) {
		return errors.New("Вы посмотрели все серии.")
	}
	a.EpCtx.Cur++
	return nil
}

func (a *Anime) SelectPreviousEp() error {
	if a.EpCtx.Cur-1 < 0 {
		return errors.New("Неверное значение курсора")
	}
	a.EpCtx.Cur--
	return nil
}

func (a *Anime) UpdateSortedEpisodeKeys() {
	var episodeKeys []int

	keys := make([]int, 0, len(a.EpCtx.Eps))
	for k := range a.EpCtx.Eps {
		keys = append(keys, k)
	}
	sort.Sort(sort.IntSlice(keys))

	for _, k := range keys {
		episodeKeys = append(episodeKeys, k)
	}

	a.EpCtx.EpsSortedKeys = episodeKeys
}
