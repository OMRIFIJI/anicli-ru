package app

import (
	"anicliru/internal/api"
	"anicliru/internal/cli/loading"
	promptsearch "anicliru/internal/cli/prompt/search"
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
	if err := a.defaultAppPipe(); err != nil {
		return err
	}
	return nil
}

func (a *App) defaultAppPipe() error {
	a.getTitleFromUser()

	if err := a.findAnime(); err != nil {
		return err
	}

	return nil
}

func (a *App) getTitleFromUser() {
	searchInput, err := promptsearch.PromptSearchInput()
	if err != nil {
		panic(err)
	}
	a.searchInput = searchInput
}

func (a *App) startLoading() {
	a.wg.Add(1)
	go loading.DisplayLoading(a.quitChan, a.wg)
}

func (a *App) stopLoading() {
	a.quitChan <- true
	a.wg.Wait()
}

func (a *App) findAnime() error {
	a.startLoading()

	err := a.api.FindAnimesByTitle(a.searchInput)
	a.stopLoading()

	return err
}
