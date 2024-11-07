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

	appCon := AppController{}

	titleName, err := appCon.PromptAnimeTitleInput()
	if err != nil {
		return err
	}

	foundAnimeInfo, err := anilibriaAPI.SearchTitleByName(titleName)
	if err != nil {
		return err
	}

	if len(foundAnimeInfo.List) == 0 {
		appCon.SearchResEmptyNotify()
		return nil
	}

	appCon.TitleSelect = promptselect.PromptSelect{
		PromptMessage: "Выберите аниме из списка:",
	}
	decoratedAnimeTitles := strdec.DecoratedAnimeTitles(foundAnimeInfo.List)
	isExitOnQuit := appCon.TitleSelect.Prompt(decoratedAnimeTitles)
	if isExitOnQuit {
		return nil
	}
	indTitle := appCon.TitleSelect.Cur.Pos
	episodes := foundAnimeInfo.List[indTitle].Media.Episodes

	appCon.EpisodeSelect = promptselect.PromptSelect{
		PromptMessage: "Выберите серию:",
	}
	episodesSlice := strdec.EpisodesToStrList(episodes)
	isExitOnQuit = appCon.EpisodeSelect.Prompt(episodesSlice)
	if isExitOnQuit {
		return nil
	}
	// Тут надо бы поправить, лучше по значения map делать все-таки
	cursorEpisode := appCon.EpisodeSelect.Cur.Pos + 1
	episodeLinks := foundAnimeInfo.GetLinks(indTitle)
    
    appCon.WatchMenuSpin(episodeLinks, cursorEpisode)

	return nil
}
