package app

import (
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
	a.quitChan <- true
	a.wg.Wait()
}

func (a *App) findAnime() error {
	a.startLoading()
	defer a.stopLoading()

	err := a.api.FindAnimesByTitle(a.searchInput)

	return err
}

func (a *App) selectAnime() error {
	animeEntries := a.api.GetAnimeTitlesWrapped()
    promptMessage := "Выберите аниме из списка:"
    prompt, err := promptselect.NewPrompt(animeEntries, promptMessage)
    if err != nil {
        return err
    }
	_, _, err = prompt.SpinPrompt()
    if err != nil {
        return err
    }

    return nil
}
