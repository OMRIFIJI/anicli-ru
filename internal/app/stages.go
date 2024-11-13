package app

import (
	"anicliru/internal/cli/loading"
	promptsearch "anicliru/internal/cli/prompt/search"
)

func (a *App) getTitleFromUser() {
	searchInput, err := promptsearch.PromptSearchInput()
	if err != nil {
		panic(err)
	}
	a.searchInput = searchInput
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

	err := a.api.FindAnimesByTitle(a.searchInput)
	a.stopLoading()

	return err
}
