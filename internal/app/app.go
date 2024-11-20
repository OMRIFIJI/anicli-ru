package app

import (
	"anicliru/internal/api"
	apilog "anicliru/internal/api/log"
	"sync"
)

type App struct {
	searchInput string
	quitChan    chan struct{}
	wg          *sync.WaitGroup
}

func NewApp() *App {
	a := App{}
	a.init()
	return &a
}

func (a *App) init() {
	apilog.Init()
	a.quitChan = make(chan struct{})
	a.wg = &sync.WaitGroup{}
}

func (a *App) RunApp() error {
    defer a.restoreTerminal()
	apilog.Init()

	if err := a.defaultAppPipe(); err != nil {
		return err
	}
	return nil
}

func (a *App) defaultAppPipe() error {
	if err := a.getTitleFromUser(); err != nil {
		return err
	}

	animes, err := a.findAnimes()
	if err != nil {
		return err
	}

	anime, isExitOnQuit, err := a.selectAnime(animes)
	if err != nil {
		return err
	}
	if isExitOnQuit {
		return nil
	}
    animes = nil

    if err := api.FindEpisodesLinks(anime); err != nil {
        return err
    }

	episode, isExitOnQuit, err := a.selectEpisode(anime)
	if err != nil {
		return err
	}
	if isExitOnQuit {
		return nil
	}

	println(anime.Title, ", ", episode.Id)

	return nil
}
