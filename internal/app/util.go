package app

import (
	"github.com/OMRIFIJI/anicli-ru/internal/api/models"
	promptselect "github.com/OMRIFIJI/anicli-ru/internal/cli/prompt/select"
)

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
