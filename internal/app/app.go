package app

import (
	"anicliru/internal/api"
	"anicliru/internal/cli/clicontroller"
	"anicliru/internal/cli/promptselect"
	"anicliru/internal/cli/strdec"
	"os/exec"
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
	isExitOnQuit := cliHand.TitleSelect.Prompt(decoratedAnimeTitles)
	if isExitOnQuit {
		return nil
	}
	indTitle := cliHand.TitleSelect.Cur.Pos
	episodes := foundAnimeInfo.List[indTitle].Media.Episodes

	cliHand.EpisodeSelect = promptselect.PromptSelect{
		PromptMessage: "Выберите серию:",
	}
	episodesList := strdec.EpisodesToStrList(episodes)
	isExitOnQuit = cliHand.EpisodeSelect.Prompt(episodesList)
	if isExitOnQuit {
		return nil
	}
    // Тут надо бы поправить
	cursorEpisode := cliHand.EpisodeSelect.Cur.Pos
	keyEpisode := episodesList[cursorEpisode]

	url := foundAnimeInfo.GetLink(indTitle, keyEpisode, "FHD")
	cmd := exec.Command("mpv", url)

	if err := cmd.Start(); err != nil {
		return err
	}


	return nil
}
