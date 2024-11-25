package app

import (
	"anicliru/internal/api"
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/models"
	"anicliru/internal/cli/loading"
	promptsearch "anicliru/internal/cli/prompt/search"
	promptselect "anicliru/internal/cli/prompt/select"
	"anicliru/internal/fmt"
)

func (a *App) getTitleFromUser() error {
	searchInput, err := promptsearch.PromptSearchInput()
	if err != nil {
		return err
	}
	a.searchInput = searchInput
	return nil
}

func (a *App) startLoading() {
	a.wg.Add(1)
	go loading.DisplayLoading(a.quitChan, a.wg)
}

func (a *App) stopLoading() {
	a.quitChan <- struct{}{}
	a.wg.Wait()
}

func (a *App) findAnimes() ([]models.Anime, error) {
	a.startLoading()
	defer a.stopLoading()

	animes, err := api.GetAnimesByTitle(a.searchInput)
	return animes, err
}

func (a *App) selectAnime(animes []models.Anime) (*models.Anime, bool, error) {
	animeEntries := entryfmt.GetWrappedAnimeTitles(animes)
	promptMessage := "Выберите аниме из списка:"

	prompt, err := promptselect.NewPrompt(animeEntries, promptMessage, true)
	if err != nil {
		return nil, false, err
	}

	isExitOnQuit, cur, err := prompt.SpinPrompt()
	if err != nil {
		return nil, false, err
	}

	animes[cur].UpdateSortedEpisodeKeys()
	return &animes[cur], isExitOnQuit, err
}

func (a *App) selectEpisode(anime *models.Anime) (bool, error) {
	episodeEntries := entryfmt.EpisodeKeysToStr(anime.EpCtx.EpsSortedKeys)
	promptMessage := "Выберите серию. " + anime.Title

	prompt, err := promptselect.NewPrompt(episodeEntries, promptMessage, false)
	if err != nil {
		return false, err
	}

	isExitOnQuit, cur, err := prompt.SpinPrompt()
	if err != nil {
		return false, err
	}

	anime.SetCur(cur)
	return isExitOnQuit, err
}

func (a *App) spinWatch(anime *models.Anime) error {
	converter := api.NewPlayerLinkConverter()

	ep, _ := anime.GetSelectedEp()
	api.GetEmbedLink(ep)

	videoLink, err := converter.GetVideoLink(ep.EmbedLink)
	if err != nil {
		return err
	}

	apilog.WarnLog.Println("Выбран эпизод")
	for key, val := range videoLink {
		apilog.WarnLog.Print(key, val)
	}

	return nil
}
