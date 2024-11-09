package app

import (
	"anicliru/internal/api"
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

    var isExitOnQuit bool
    var cursor int

	decoratedAnimeTitles := strdec.DecoratedAnimeTitles(foundAnimeInfo.List)
	isExitOnQuit, cursor = appCon.TitleSelect.NewPrompt(decoratedAnimeTitles, "Выберите аниме из списка:")
	if isExitOnQuit {
		return nil
	}
    cursorTitle := cursor
	episodes := foundAnimeInfo.List[cursor].Media.Episodes

	episodesSlice := strdec.EpisodesToStrList(episodes)
	isExitOnQuit, cursor = appCon.EpisodeSelect.NewPrompt(episodesSlice, "Выберите серию:")
	if isExitOnQuit {
		return nil
	}
	// Тут надо бы поправить, лучше по значения map делать все-таки
	cursorEpisode := cursor + 1
	episodeLinks := foundAnimeInfo.GetLinks(cursorTitle)
    
    appCon.WatchMenuSpin(episodeLinks, cursorEpisode)

	return nil
}
