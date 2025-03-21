package models

import (
	"errors"
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

type EpisodesContext struct {
	Cur          int             `json:"cur"`
	Eps          map[int]Episode `json:"-"`
	TotalEpCount int             `json:"-"`
	AiredEpCount int             `json:"-"`
}

type Episode struct {
	Id         int
	EmbedLinks EmbedLinks
}

// Структура аниме, возвращаемая api.
type Anime struct {
	Id        int             `json:"id"`
	Uname     string          `json:"uname,omitempty"`
	Provider  string          `json:"provider"`
	EpCtx     EpisodesContext `json:"epCtx"`
	Title     string          `json:"-"`
	MediaType string          `json:"-"`
	SearchPos int             `json:"-"`
}

func (e *EpisodesContext) GetSelectedEp() (*Episode, error) {
	ep, exists := e.Eps[e.Cur]
	if !exists {
		return nil, errors.New("выбранный эпизод не существует")
	}
	return &ep, nil
}

func (e *EpisodesContext) SetCur(cur int) error {
	if cur < 1 || cur > e.AiredEpCount {
		return errors.New("неверное значение курсора")
	}
	e.Cur = cur
	return nil
}

func (e *EpisodesContext) SelectNextEp() error {
	if e.Cur+1 > e.AiredEpCount {
		return errors.New("вы посмотрели все серии")
	}
	e.Cur++
	return nil
}

func (e *EpisodesContext) SelectPreviousEp() error {
	if e.Cur-1 < 1 {
		return errors.New("неверное значение курсора")
	}
	e.Cur--
	return nil
}
