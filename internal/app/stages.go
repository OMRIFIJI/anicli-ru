package app

import (
	"anicliru/internal/api/models"
	promptsearch "anicliru/internal/cli/prompt/search"
	promptselect "anicliru/internal/cli/prompt/select"
	entryfmt "anicliru/internal/fmt"
	"anicliru/internal/logger"
	"sync"
)

func (a *App) getTitleFromUser() (string, error) {
	searchInput, err := promptsearch.PromptSearchInput()
	if err != nil {
		return "", err
	}
	return searchInput, nil
}

func (a *App) findAnimes(searchInput string) ([]models.Anime, error) {
	a.startLoading()
	defer a.stopLoading()

	animes, err := a.api.GetAnimesByTitle(searchInput)
	return animes, err
}

func (a *App) selectAnime(animes []models.Anime) (*models.Anime, bool, error) {
	animeEntries := entryfmt.WrapAnimeTitlesAired(animes)
	cur, isExitOnQuit, err := promptAnime(animes, animeEntries)
	if err != nil {
		return nil, false, err
	}

	a.api.SetAllEmbedLinks(&animes[cur])
	return &animes[cur], isExitOnQuit, err
}

func (a *App) selectAnimeToCountinue(animes []models.Anime) (*models.Anime, bool, error) {
	animeEntries := entryfmt.WrapAnimeTitlesWatched(animes)
	cur, isExitOnQuit, err := promptAnime(animes, animeEntries)
	if err != nil {
		return nil, false, err
	}

	a.api.SetAllEmbedLinks(&animes[cur])
	return &animes[cur], isExitOnQuit, err
}

func (a *App) selectEpisode(anime *models.Anime) (bool, error) {
	episodeEntries := entryfmt.EpisodeEntries(anime.EpCtx)
	promptMessage := "Выберите серию. " + anime.Title

	prompt, err := promptselect.NewPrompt(episodeEntries, promptMessage, false)
	if err != nil {
		return false, err
	}

	isExitOnQuit, cur, err := prompt.SpinPrompt()
	if err != nil {
		return false, err
	}

	err = anime.EpCtx.SetCur(cur + 1)
	if err != nil {
		return false, err
	}
	return isExitOnQuit, nil
}

func (a *App) prepareSavedAnime(animeSlice []models.Anime) ([]models.Anime, error) {
	var animeSlicePrepared []models.Anime
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, anime := range animeSlice {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.api.PrepareSavedAnime(&anime); err != nil {
				logger.ErrorLog.Println(err)
			}
			// Если онгоинг и новых серий нет, то не выводим
			if anime.EpCtx.Cur != anime.EpCtx.AiredEpCount {
				mu.Lock()
				defer mu.Unlock()
				animeSlicePrepared = append(animeSlicePrepared, anime)
			}
		}()
	}
	wg.Wait()

	return animeSlicePrepared, nil
}
