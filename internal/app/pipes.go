package app

import (
	"anicliru/internal/api"
	config "anicliru/internal/app/cfg"
	promptselect "anicliru/internal/cli/prompt/select"
	"anicliru/internal/db"
	"anicliru/internal/video"
	"errors"
	"fmt"
)

func defaultPipe(dbh *db.DBHandler) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	api, err := initApi(dbh, cfg)
	if err != nil {
		return err
	}

	searchInput, err := getTitleFromUser()
	if err != nil {
		return err
	}

	animes, err := findAnimes(searchInput, api)
	if err != nil {
		return err
	}

	oldTermState, err := promptselect.PrepareTerminal()
	if err != nil {
		return err
	}
	defer promptselect.RestoreTerminal(oldTermState)

	anime, isExitOnQuit, err := selectAnime(animes, api)
	if err != nil {
		return err
	}
	if isExitOnQuit {
		return nil
	}

	animeFromDB, err := dbh.GetAnime(anime.Title)
	// Если ещё не смотрели аниме или не удалось загрузить, то выбираем эпизод
	if err != nil {
		isExitOnQuit, err = selectEpisode(anime)
		if err != nil {
			return err
		}
		if isExitOnQuit {
			return nil
		}
	} else {
		if animeFromDB.EpCtx.Cur == anime.EpCtx.AiredEpCount {
			return errors.New("вы просмотрели все доступные серии")
		}
		anime.EpCtx.Cur = animeFromDB.EpCtx.Cur
	}

	// Сохраняет информацию об аниме на выходе
	defer dbh.UpdateAnime(anime)

	animePlayer := video.NewAnimePlayer(anime, api, &cfg.Video)
	return animePlayer.Play()
}

func continuePipe(dbh *db.DBHandler) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	api, err := initApi(dbh, cfg)
	if err != nil {
		return err
	}

	animes, err := dbh.GetAnimeSlice()
	if err != nil {
		return err
	}

	animes, err = prepareSavedAnime(animes, api)
	if err != nil {
		return err
	}

	if len(animes) == 0 {
		return errors.New("нет аниме для продолжения просмотра")
	}

	oldTermState, err := promptselect.PrepareTerminal()
	if err != nil {
		return err
	}
	defer promptselect.RestoreTerminal(oldTermState)

	anime, isExitOnQuit, err := selectAnimeWithState(animes, api)
	if err != nil {
		return err
	}
	if isExitOnQuit {
		return nil
	}

	// Если источник больше не отвечает, ищем аниме заново во всех источниках.
	if anime.Provider == "" {
		dbCur := anime.EpCtx.Cur
		animes, err := api.GetAnimesByTitle(anime.Title)
		if err != nil {
			return err
		}

		anime, isExitOnQuit, err = selectAnime(animes, api)
		if err != nil {
			return err
		}
		if isExitOnQuit {
			return nil
		}

		anime.EpCtx.Cur = dbCur
	}

	defer dbh.UpdateAnime(anime)

	animePlayer := video.NewAnimePlayer(anime, api, &cfg.Video)
	return animePlayer.Play()
}

func deletePipe(dbh *db.DBHandler) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	api, err := initApi(dbh, cfg)
	if err != nil {
		return err
	}

	animes, err := dbh.GetAnimeSlice()
	if err != nil {
		return err
	}

	animes, err = prepareSavedAnime(animes, api)
	if err != nil {
		return err
	}

	if len(animes) == 0 {
		return errors.New("нет аниме для удаления")
	}

	oldTermState, err := promptselect.PrepareTerminal()
	if err != nil {
		return err
	}
	defer promptselect.RestoreTerminal(oldTermState)

	anime, isExitOnQuit, err := selectAnimeWithState(animes, api)
	if err != nil {
		return err
	}
	if isExitOnQuit {
		return nil
	}

	dbh.DeleteAnime(anime.Title)
	return nil
}

func deleteAllPipe(dbh *db.DBHandler) error {
	if err := dbh.DeleteAllAnime(); err != nil {
		return err
	}

	return nil
}

func checkProvidersPipe() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	fmt.Print(api.GetProvidersState(cfg.Providers))

	return nil
}
