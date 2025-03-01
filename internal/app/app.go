package app

import (
	"anicliru/internal/api"
	config "anicliru/internal/app/cfg"
	promptselect "anicliru/internal/cli/prompt/select"
	"anicliru/internal/logger"
	"anicliru/internal/video"
	"sync"
)

type App struct {
	api      *api.AnimeAPI
	quitChan chan struct{}
	wg       *sync.WaitGroup
}

func NewApp() (*App, error) {
	a := App{}
    if err := a.init(); err != nil {
        return nil, err
    }
	return &a, nil
}

func (a *App) init() error {
    cfg, err := config.LoadConfig()
    if err != nil {
        return err
    }

    api, err := api.NewAnimeAPI(cfg.Providers)
    if err != nil {
        return err
    }

    a.api = api
	a.quitChan = make(chan struct{})
	a.wg = &sync.WaitGroup{}

	logger.Init()

    return nil
}

func (a *App) RunApp() error {

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

	animePlayer := video.NewAnimePlayer(anime, a.api)
	if err := animePlayer.Play(); err != nil {
		return err
	}

	return nil
}
