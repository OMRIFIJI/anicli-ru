package app

import (
	"anicliru/internal/api/models"
	"anicliru/internal/cli/loading"
	promptsearch "anicliru/internal/cli/prompt/search"
	promptselect "anicliru/internal/cli/prompt/select"
	entryfmt "anicliru/internal/fmt"
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
	var wg sync.WaitGroup
	quitChan := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		loading.DisplayLoading(quitChan)
	}()

	defer func() {
		defer loading.RestoreTerminal()
		quitChan <- struct{}{}
		wg.Wait()
	}()

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

func (a *App) selectAnimeWithState(animes []models.Anime) (*models.Anime, bool, error) {
	animeEntries := entryfmt.WrapAnimeTitlesWatched(animes)
	cur, isExitOnQuit, err := promptAnime(animes, animeEntries)
	if err != nil {
		return nil, false, err
	}

    // Если источник доступен, то заполняем эмбеды
    if animes[cur].Provider != "" {
        a.api.SetAllEmbedLinks(&animes[cur])
    }
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
	var mu sync.Mutex
	var wg sync.WaitGroup

	var animeSlicePrepared []models.Anime
	for _, anime := range animeSlice {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.api.PrepareSavedAnime(&anime)
			// Если удалось загрузить и новых серий нет, то не выводим
			if anime.EpCtx.Cur == anime.EpCtx.AiredEpCount && err == nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()
			animeSlicePrepared = append(animeSlicePrepared, anime)

		}()
	}
	wg.Wait()

	return animeSlicePrepared, nil
}
