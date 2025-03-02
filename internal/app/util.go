package app

import (
	"anicliru/internal/api/models"
	"anicliru/internal/cli/loading"
	promptselect "anicliru/internal/cli/prompt/select"
)

func (a *App) startLoading() {
	a.wg.Add(1)
	go loading.DisplayLoading(a.quitChan, a.wg)
}

func (a *App) stopLoading() {
	defer loading.RestoreTerminal()
	a.quitChan <- struct{}{}
	a.wg.Wait()
}

func promptAnime(animes []models.Anime, entries []string) (int, bool, error) {
	promptMessage := "Выберите аниме из списка:"

	prompt, err := promptselect.NewPrompt(entries, promptMessage, true)
	if err != nil {
		return 0, false, err
	}

	isExitOnQuit, cur, err := prompt.SpinPrompt()
	if err != nil {
		return 0, false, err
	}

	return cur, isExitOnQuit, err
}
