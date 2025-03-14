package app

import (
	"anicliru/internal/api"
	"anicliru/internal/api/models"
	config "anicliru/internal/app/cfg"
	"anicliru/internal/cli/loading"
	promptsearch "anicliru/internal/cli/prompt/search"
	promptselect "anicliru/internal/cli/prompt/select"
	"anicliru/internal/db"
	entryfmt "anicliru/internal/fmt"
	"sync"
)

func initApi(dbh *db.DBHandler, cfg *config.Config) (*api.AnimeAPI, error) {
	api, err := api.NewAnimeAPI(cfg.Providers.DomainMap, cfg.Players.Domains, dbh)
	if err != nil {
		return nil, err
	}

	return api, nil
}

func getTitleFromUser() (string, error) {
	searchInput, err := promptsearch.PromptSearchInput()
	if err != nil {
		return "", err
	}
	return searchInput, nil
}

func findAnimes(searchInput string, api *api.AnimeAPI) ([]models.Anime, error) {
	var wg sync.WaitGroup
	quitChan := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		loading.DisplayLoading(quitChan)
	}()

	defer func() {
		defer loading.RestoreTerminal()
		quitChan <- struct{}{}
		wg.Wait()
	}()

	animes, err := api.GetAnimesByTitle(searchInput)
	return animes, err
}

func selectAnime(animes []models.Anime, api *api.AnimeAPI) (*models.Anime, bool, error) {
	animeEntries := entryfmt.WrapAnimeTitlesAired(animes)
	cur, isExitOnQuit, err := promptAnime(animes, animeEntries)
	if err != nil {
		return nil, false, err
	}

	api.SetAllEmbedLinks(&animes[cur])
	return &animes[cur], isExitOnQuit, err
}

func selectAnimeWithState(animes []models.Anime, api *api.AnimeAPI) (*models.Anime, bool, error) {
	animeEntries := entryfmt.WrapAnimeTitlesWatched(animes)
	cur, isExitOnQuit, err := promptAnime(animes, animeEntries)
	if err != nil {
		return nil, false, err
	}

	// Если источник доступен, то заполняем эмбеды
	if animes[cur].Provider != "" {
		api.SetAllEmbedLinks(&animes[cur])
	}
	return &animes[cur], isExitOnQuit, err
}

func selectEpisode(anime *models.Anime) (bool, error) {
	episodeEntries := entryfmt.EpisodeEntries(anime.EpCtx)
	promptMessage := "Выберите серию. " + anime.Title

	prompt, err := promptselect.NewPrompt(episodeEntries, promptMessage, false)
	if err != nil {
		return false, err
	}

	isExitOnQuit, cur, err := prompt.SpinPrompt()
	if err != nil {
		return false, err
	}

	err = anime.EpCtx.SetCur(cur + 1)
	if err != nil {
		return false, err
	}
	return isExitOnQuit, nil
}

func prepareSavedAnime(animes []models.Anime, api *api.AnimeAPI) ([]models.Anime, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup

	var animesPrepared []models.Anime
	for _, anime := range animes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := api.PrepareSavedAnime(&anime)
			// Если удалось загрузить и новых серий нет, то не выводим
			if anime.EpCtx.Cur == anime.EpCtx.AiredEpCount && err == nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()
			animesPrepared = append(animesPrepared, anime)

		}()
	}
	wg.Wait()

	return animesPrepared, nil
}
