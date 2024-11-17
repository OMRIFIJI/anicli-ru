package app

import (
	"anicliru/internal/animefmt"
	"anicliru/internal/api/types"
	"anicliru/internal/cli/loading"
	promptsearch "anicliru/internal/cli/prompt/search"
	promptselect "anicliru/internal/cli/prompt/select"
)

func (a *App) getTitleFromUser() error {
	searchInput, err := promptsearch.PromptSearchInput()
	if err != nil {
		return err
	}
	a.searchInput = searchInput
	return nil
}

func (a *App) startLoading() {
	a.wg.Add(1)
	go loading.DisplayLoading(a.quitChan, a.wg)
}

func (a *App) stopLoading() {
	a.quitChan <- struct{}{}
	a.wg.Wait()
}

func (a *App) findAnimes() ([]types.Anime, error) {
	a.startLoading()
	defer a.stopLoading()

	animes, err := a.api.FindAnimesByTitle(a.searchInput)
	return animes, err
}

func (a *App) selectAnime(animes []types.Anime) (*types.Anime, bool, error) {
	animeEntries := animefmt.GetWrappedAnimeTitles(animes)
	promptMessage := "Выберите аниме из списка:"

	prompt, err := promptselect.NewPrompt(animeEntries, promptMessage, true)
	if err != nil {
		return nil, false, err
	}

	isExitOnQuit, cur, err := prompt.SpinPrompt()
	if err != nil {
		return nil, false, err
	}

	return &animes[cur], isExitOnQuit, err
}
