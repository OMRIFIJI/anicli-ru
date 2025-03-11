package app

import (
	promptselect "anicliru/internal/cli/prompt/select"
	"anicliru/internal/db"
	"anicliru/internal/video"
	"errors"
)

func (a *App) defaultPipe(dbh *db.DBHandler) error {
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

	animeFromDB, err := dbh.GetAnime(anime.Title)
	// Если ещё не смотрели аниме или не удалось загрузить, то выбираем эпизод
	if err != nil {
		isExitOnQuit, err = a.selectEpisode(anime)
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

	animePlayer := video.NewAnimePlayer(anime, a.api)
	if err := animePlayer.Play(); err != nil {
		return err
	}

	return nil
}

func (a *App) continuePipe(dbh *db.DBHandler) error {
	animeSlice, err := dbh.GetAnimeSlice()
	if err != nil {
		return err
	}

	animes, err := a.prepareSavedAnime(animeSlice)
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

	anime, isExitOnQuit, err := a.selectAnimeWithState(animes)
	if err != nil {
		return err
	}
	if isExitOnQuit {
		return nil
	}

	// Если источник больше не отвечает, ищем аниме заново во всех источниках.
	if anime.Provider == "" {
		dbCur := anime.EpCtx.Cur
		animes, err := a.api.GetAnimesByTitle(anime.Title)
		if err != nil {
			return err
		}

		anime, isExitOnQuit, err = a.selectAnime(animes)
		if err != nil {
			return err
		}
		if isExitOnQuit {
			return nil
		}

		anime.EpCtx.Cur = dbCur
	}

	defer dbh.UpdateAnime(anime)

	animePlayer := video.NewAnimePlayer(anime, a.api)
	if err := animePlayer.Play(); err != nil {
		return err
	}

	return nil
}

func (a *App) deletePipe(dbh *db.DBHandler) error {
	animes, err := dbh.GetAnimeSlice()
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

	anime, isExitOnQuit, err := a.selectAnimeWithState(animes)
	if err != nil {
		return err
	}
	if isExitOnQuit {
		return nil
	}

	dbh.DeleteAnime(anime.Title)
	return nil
}

func (a *App) deleteAllPipe(dbh *db.DBHandler) error {
	if err := dbh.DeleteAllAnime(); err != nil {
		return err
	}

	return nil
}
