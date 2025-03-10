package app

import (
	"anicliru/internal/api"
	config "anicliru/internal/app/cfg"
	"anicliru/internal/db"
	"anicliru/internal/logger"
	"flag"
	"fmt"

	"rsc.io/getopt"
)

type App struct {
	api *api.AnimeAPI
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
	deletePtr := flag.Bool("delete", false, "удалить запись из базы данных, просматриваемых аниме")
	deleteAllPtr := flag.Bool("delete-all", false, "удалить все записи из базы данных, просматриваемых аниме")

	getopt.Aliases(
		"c", "continue",
		"h", "help",
		"d", "delete",
	)
	getopt.Parse()

	if *helpPtr {
		flag.Usage()
		return nil
	}

	if *continuePtr {
		if err := a.continuePipe(dbh); err != nil {
			return err
		}
		return nil
	}

	if *deletePtr {
		if err := a.deletePipe(dbh); err != nil {
			return err
		}
		return nil
	}

    if *deleteAllPtr {
		if err := a.deleteAllPipe(dbh); err != nil {
			return err
		}
		return nil
    }

	if flag.Parsed() {
		if err := a.defaultPipe(dbh); err != nil {
			return err
		}
	}
	return nil
}
