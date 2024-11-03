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

    cli := CLI{
        cursor: 0,
    }

	titleName, err := cli.PromptAnimeTitle()
	if err != nil {
		return err
	}

	foundAnimeInfo, err := anilibriaAPI.SearchTitleByName(titleName)
	if err != nil {
		return err
	}

	if len(foundAnimeInfo.List) == 0 {
        cli.SearchResEmptyNotify()
		return nil
	}

    cli.PromptSearchRes(foundAnimeInfo.List)

	return nil
}
