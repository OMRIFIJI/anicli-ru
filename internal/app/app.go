package app

import (
	apilog "anicliru/internal/api/log"
	promptselect "anicliru/internal/cli/prompt/select"
	"anicliru/internal/video"
	"sync"
)

type App struct {
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
	apilog.Init()

	if err := a.defaultAppPipe(); err != nil {
		return err
	}
	return nil
}

func (a *App) defaultAppPipe() error {
	searchInput, err := a.getTitleFromUser()
    if err != nil {
		return err
	}

	animes, err := a.findAnimes(searchInput)
	if err != nil {
		return err
	}

	oldTermState, err := promptselect.PrepareTerminal()
	if err != nil {
		return err
	}
	defer promptselect.RestoreTerminal(oldTermState)

	anime, isExitOnQuit, err := a.selectAnime(animes)
	if err != nil {
		return err
	}
	if isExitOnQuit {
		return nil
	}
	animes = nil

	isExitOnQuit, err = a.selectEpisode(anime)
	if err != nil {
		return err
	}
	if isExitOnQuit {
		return nil
	}

    animePlayer := video.NewAnimePlayer(anime)
    if err := animePlayer.SpinWatch(); err != nil{
        return err
    }

	return nil
}
