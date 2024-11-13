package app

import (
	"anicliru/internal/api"
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/types"
	"anicliru/internal/cli/prompt/select"
	"errors"
	"fmt"
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

	if err := a.defaultAppPipe(); err != nil {
        var animeError *types.AnimeError
        // Не удалось обработать часть аниме
        if errors.As(err, &animeError){
            fmt.Print(err)
        } else {
            return err
        }
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
