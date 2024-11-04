package cli

import (
	"anicliru/internal/api"
)

func StartApp() error {
	client := api.InitHttpClient()
	anilibriaAPI := api.AnilibriaAPI{
		Client:       client,
		BaseURL:      "https://api.anilibria.tv/v3/",
		SearchMethod: "title/search",
	}

	cliHand := CLIHandler{}

	titleName, err := cliHand.PromptAnimeTitle()
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

	decoratedAnimeTitles := decoratedAnimeTitles(foundAnimeInfo.List)

	cliHand.titleSelect = BaseSelectHandler{
		cursor:         0,
		foundAnimeInfo: decoratedAnimeTitles,
	}

	cliHand.titleSelect.PromptSearchRes(decoratedAnimeTitles)

	return nil
}
