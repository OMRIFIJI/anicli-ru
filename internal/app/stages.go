package app

import (
	"sync"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animefmt"
	"github.com/OMRIFIJI/anicli-ru/internal/app/config"
	"github.com/OMRIFIJI/anicli-ru/internal/cli/loading"
	promptsearch "github.com/OMRIFIJI/anicli-ru/internal/cli/prompt/search"
	promptselect "github.com/OMRIFIJI/anicli-ru/internal/cli/prompt/select"
	"github.com/OMRIFIJI/anicli-ru/internal/db"
)

func initApi(dbh *db.DBHandler, cfg *config.Config) (*animeapi.AnimeAPI, error) {
	api, err := animeapi.NewAnimeAPI(cfg.Providers.DomainMap, cfg.Players.Domains, dbh)
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

func findAnimes(searchInput string, api *animeapi.AnimeAPI) ([]models.Anime, error) {
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

func selectAnime(animes []models.Anime, api *animeapi.AnimeAPI) (*models.Anime, bool, error) {
	animeEntries := animefmt.WrapAnimeTitlesAired(animes)
	cur, isExitOnQuit, err := promptAnime(animeEntries)
	if err != nil {
		return nil, false, err
	}

	api.SetAllEmbedLinks(&animes[cur])
	return &animes[cur], isExitOnQuit, err
}

func selectAnimeWithState(animes []models.Anime, api *animeapi.AnimeAPI) (*models.Anime, bool, error) {
	animeEntries := animefmt.WrapAnimeTitlesWatched(animes)
	cur, isExitOnQuit, err := promptAnime(animeEntries)
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
	episodeEntries := animefmt.EpisodeEntries(anime.EpCtx)
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

func prepareSavedAnime(animes []models.Anime, api *animeapi.AnimeAPI) ([]models.Anime, error) {
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
