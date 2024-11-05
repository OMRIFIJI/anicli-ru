package app

import (
	"anicliru/internal/api"
	"anicliru/internal/cli/clicontroller"
	"anicliru/internal/cli/promptselect"
	"anicliru/internal/cli/strdec"
)

func StartApp() error {
	client := api.InitHttpClient()
	anilibriaAPI := api.AnilibriaAPI{
		Client:       client,
		BaseURL:      "https://api.anilibria.tv/v3/",
		SearchMethod: "title/search",
	}

	cliHand := clicontroller.CLIController{}

	titleName, err := cliHand.PromptAnimeTitleInput()
	if err != nil {
		return err
	}

	foundAnimeInfo, err := anilibriaAPI.SearchTitleByName(titleName)
	if err != nil {
		return err
	}

	if len(foundAnimeInfo.List) == 0 {
		cliHand.SearchResEmptyNotify()
		return nil
	}

	cliHand.TitleSelect = promptselect.PromptSelect{
        PromptMessage: "Выберите аниме из списка:",
    }
	decoratedAnimeTitles := strdec.DecoratedAnimeTitles(foundAnimeInfo.List)
	// true если пользователь вышел через "q"
	isExitOnQuit := cliHand.TitleSelect.Prompt(decoratedAnimeTitles)
	if isExitOnQuit {
		return nil
	}
	cursor := cliHand.TitleSelect.Cur.Pos
	episodes := foundAnimeInfo.List[cursor].Media.Episodes

	cliHand.EpisodeSelect = promptselect.PromptSelect{
        PromptMessage: "Выберите серию:",
    }
	episodesList := strdec.EpisodesToStrList(episodes)
	isExitOnQuit = cliHand.EpisodeSelect.Prompt(episodesList)
	if isExitOnQuit {
		return nil
	}

	return nil
}
