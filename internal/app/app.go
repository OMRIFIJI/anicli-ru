package app

import (
	"anicliru/internal/api"
	"anicliru/internal/cli/promptselect"
	"anicliru/internal/cli/strdec"
)

func RunApp() error {
	client := api.InitHttpClient()
	anilibriaAPI := api.AnilibriaAPI{
		Client:       client,
		BaseURL:      "https://api.anilibria.tv/v3/",
		SearchMethod: "title/search",
	}

	cliCon := AppController{}

	titleName, err := cliCon.PromptAnimeTitleInput()
	if err != nil {
		return err
	}

	foundAnimeInfo, err := anilibriaAPI.SearchTitleByName(titleName)
	if err != nil {
		return err
	}

	if len(foundAnimeInfo.List) == 0 {
		cliCon.SearchResEmptyNotify()
		return nil
	}

	cliCon.TitleSelect = promptselect.PromptSelect{
		PromptMessage: "Выберите аниме из списка:",
	}
	decoratedAnimeTitles := strdec.DecoratedAnimeTitles(foundAnimeInfo.List)
	isExitOnQuit := cliCon.TitleSelect.Prompt(decoratedAnimeTitles)
	if isExitOnQuit {
		return nil
	}
	indTitle := cliCon.TitleSelect.Cur.Pos
	episodes := foundAnimeInfo.List[indTitle].Media.Episodes

	cliCon.EpisodeSelect = promptselect.PromptSelect{
		PromptMessage: "Выберите серию:",
	}
	episodesSlice := strdec.EpisodesToStrList(episodes)
	isExitOnQuit = cliCon.EpisodeSelect.Prompt(episodesSlice)
	if isExitOnQuit {
		return nil
	}
	// Тут надо бы поправить, лучше по значения map делать все-таки
	cursorEpisode := cliCon.EpisodeSelect.Cur.Pos + 1
	episodeLinks := foundAnimeInfo.GetLinks(indTitle)
    
    cliCon.WatchMenuSpin(episodeLinks, cursorEpisode)

	return nil
}
