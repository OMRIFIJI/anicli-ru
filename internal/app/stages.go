package app

import (
	"anicliru/internal/api/models"
	"anicliru/internal/cli/loading"
	promptsearch "anicliru/internal/cli/prompt/search"
	promptselect "anicliru/internal/cli/prompt/select"
	"anicliru/internal/fmt"
)

func (a *App) getTitleFromUser() (string, error) {
	searchInput, err := promptsearch.PromptSearchInput()
	if err != nil {
		return "", err
	}
	return searchInput, nil
}

func (a *App) startLoading() {
	a.wg.Add(1)
	go loading.DisplayLoading(a.quitChan, a.wg)
}

func (a *App) stopLoading() {
	defer loading.RestoreTerminal()
	a.quitChan <- struct{}{}
	a.wg.Wait()
}

func (a *App) findAnimes(searchInput string) ([]models.Anime, error) {
	a.startLoading()
	defer a.stopLoading()

	animes, err := a.api.GetAnimesByTitle(searchInput)
	return animes, err
}

func (a *App) selectAnime(animes []models.Anime) (*models.Anime, bool, error) {
	animeEntries := entryfmt.GetWrappedAnimeTitles(animes)
	promptMessage := "Выберите аниме из списка:"

	prompt, err := promptselect.NewPrompt(animeEntries, promptMessage, true)
	if err != nil {
		return nil, false, err
	}

	isExitOnQuit, cur, err := prompt.SpinPrompt()
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
