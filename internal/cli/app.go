package cli

import (
	"anicliru/internal/api"
	"fmt"
)

func StartApp() error {
	client := api.InitHttpClient()
	anilibriaAPI := api.AnilibriaAPI{
		Client:       client,
		BaseURL:      "https://api.anilibria.tv/v3/",
		SearchMethod: "title/search",
	}

	cli := CLI{}

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

	foundAnimeTitles := make([]string, 0, len(foundAnimeInfo.List))

	for _, animeInfo := range foundAnimeInfo.List {
		episodeCount := len(animeInfo.Media.Episodes)
		var episodeSuffixStr string
		if episodeCount == 1 {
			episodeSuffixStr = fmt.Sprintf(" (%d %s)", episodeCount, "эпизод")
		} else if episodeCount > 1 {
			episodeSuffixStr = fmt.Sprintf(" (%d %s)", episodeCount, "эпизодов")
		} else {
			episodeSuffixStr = fmt.Sprintf(" (Не доступно)")
		}
		foundAnimeTitles = append(foundAnimeTitles, animeInfo.Names.RU+episodeSuffixStr)
	}

	cli.titleChoice = BaseChoiceHandler{
		cursor:         0,
		foundAnimeInfo: foundAnimeTitles,
	}

	cli.titleChoice.PromptSearchRes(foundAnimeTitles)

	return nil
}
