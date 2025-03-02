package app

import (
	"anicliru/internal/api"
	config "anicliru/internal/app/cfg"
	promptselect "anicliru/internal/cli/prompt/select"
	"anicliru/internal/db"
	"anicliru/internal/logger"
	"anicliru/internal/video"
	"flag"
	"fmt"
	"sync"

	"rsc.io/getopt"
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
	dbh, err := db.NewDBHandler()
	if err != nil {
		return err
	}
	defer dbh.CloseDB()

	flag.Usage = func() {
		fmt.Println("Доступные аргументы командной строки для anicli-ru:")
		getopt.PrintDefaults()
	}

	helpPtr := flag.Bool("help", false, "выводит сообщение, которое вы сейчас видите")
	continuePtr := flag.Bool("continue", false, "продолжить просмотр аниме")

	getopt.Aliases(
		"c", "continue",
		"h", "help",
	)
	getopt.Parse()

	if *helpPtr {
		flag.Usage()
		return nil
	}

	if *continuePtr {
		if err := a.continueAppPipe(dbh); err != nil {
			return err
		}
		return nil
	}

	if flag.Parsed() {
		if err := a.defaultAppPipe(dbh); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) continueAppPipe(dbh *db.DBHandler) error {
	animeSlice, err := dbh.GetAnimeSlice()
	if err != nil {
		return err
	}

	animes, err := a.prepareSavedAnime(animeSlice)
	if err != nil {
		return err
	}

	if len(animes) == 0 {
		fmt.Println("Нет аниме для продолжения просмотра.")
		return nil
	}

	oldTermState, err := promptselect.PrepareTerminal()
	if err != nil {
		return err
	}
	defer promptselect.RestoreTerminal(oldTermState)

	anime, isExitOnQuit, err := a.selectAnimeToCountinue(animes)
	if err != nil {
		return err
	}
	if isExitOnQuit {
		return nil
	}
	animes = nil

	defer dbh.UpdateAnime(anime)

	animePlayer := video.NewAnimePlayer(anime, a.api)
	if err := animePlayer.Play(); err != nil {
		return err
	}

	return nil
}

func (a *App) defaultAppPipe(dbh *db.DBHandler) error {
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

	// Сохраняет информацию об аниме на выходе
	defer dbh.UpdateAnime(anime)

	animePlayer := video.NewAnimePlayer(anime, a.api)
	if err := animePlayer.Play(); err != nil {
		return err
	}

	return nil
}
