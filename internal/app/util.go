package app

import (
	promptselect "github.com/OMRIFIJI/anicli-ru/internal/cli/prompt/select"
)

func promptAnime(entries []string) (int, bool, error) {
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
