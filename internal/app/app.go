package app

import (
	"anicliru/internal/api"
	"anicliru/internal/cli"
)

func StartApp() error {
	client := api.InitHttpClient()
	anilibriaAPI := api.AnilibriaAPI{
		Client:       client,
		BaseURL:      "https://api.anilibria.tv/v3/",
		SearchMethod: "title/search",
	}

	cliHand := cli.CLIHandler{}

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


	cliHand.TitleSelect = cli.SelectPrompt{}
	decoratedAnimeTitles := cli.DecoratedAnimeTitles(foundAnimeInfo.List)
	exitCode := cliHand.TitleSelect.PromptSearchRes(decoratedAnimeTitles)
	if exitCode == cli.QuitCode {
		return nil
	}
	cursor := cliHand.TitleSelect.Cur.Pos
	episodes := foundAnimeInfo.List[cursor].Media.Episodes

    cliHand.EpisodeSelect = cli.SelectPrompt{}
	episodesList := cli.EpisodesToStrList(episodes)
	exitCode = cliHand.TitleSelect.PromptSearchRes(episodesList)
	if exitCode == cli.QuitCode {
		return nil
	}

	return nil
}
