package app

import (
	"anicliru/internal/api"
	apilog "anicliru/internal/api/log"
	clilog "anicliru/internal/cli/log"
	"anicliru/internal/cli/prompt/select"
	"sync"
)

type App struct {
	searchInput string
	api         api.API
	prompt      promptselect.PromptSelect
	quitChan    chan bool
	wg          *sync.WaitGroup
}

func NewApp() *App {
	a := App{}
	a.init()
	return &a
}

func (a *App) init() {
	a.quitChan = make(chan bool)
	a.wg = &sync.WaitGroup{}
}

func (a *App) RunApp() error {
	apilog.Init()
    clilog.Init()

	if err := a.defaultAppPipe(); err != nil {
		return err
	}
	return nil
}

func (a *App) defaultAppPipe() error {
	if err := a.getTitleFromUser(); err != nil {
		return err
	}

	if err := a.findAnime(); err != nil {
		return err
	}

	return nil
}
