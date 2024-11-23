package models

import (
	"errors"
	"sort"
)

const FilmEpisodeId = -1

// Структура: озвучка -> плеер -> ссылка на embed
type EmbedLink map[string]map[string]string

// Структура: озвучка -> качество видео -> ссылки на видео (со всех плееров)
type VideoLink map[string]map[string][]string

type EpisodesContext struct {
	Eps           map[int]*Episode
	EpsSortedKeys []int
	Cur           int
	TotalEpCount  int
}

type Episode struct {
	Id        int
	EmbedLink EmbedLink
}

type Anime struct {
	Id        string
	Uname     string
	Title     string
	MediaType string
	EpCtx     EpisodesContext
}

func (a *Anime) GetSelectedEp() (*Episode, error) {
	cur := a.EpCtx.Cur
	key := a.EpCtx.EpsSortedKeys[cur]
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
		return errors.New("Неверное значение курсора")
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
