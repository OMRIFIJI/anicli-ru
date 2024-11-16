package app

import (
	"anicliru/internal/api"
	apilog "anicliru/internal/api/log"
	"sync"
)

type App struct {
	searchInput string
	api         api.API
	quitChan    chan bool
	wg          *sync.WaitGroup
}

func NewApp() *App {
	a := App{}
	a.init()
	return &a
}

func (a *App) init() {
	apilog.Init()
	a.quitChan = make(chan bool)
	a.wg = &sync.WaitGroup{}
}

func (a *App) RunApp() error {
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

	print(anime.Title)

	return nil
}
