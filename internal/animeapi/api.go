package animeapi

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/providers"
	"github.com/OMRIFIJI/anicli-ru/internal/db"
	httpkit "github.com/OMRIFIJI/anicli-ru/internal/httpkit"
	"github.com/OMRIFIJI/anicli-ru/internal/logger"
)

type API struct {
	animeParsers map[string]animeParser
	Converter    *player.PlayerLinkConverter
}

// Интерфейс, который должны реализовывать парсеры аниме-сайтов.
//
// GetAnimesByTitle - на вход принимает название или часть названия аниме title.
// По этому значению находит несколько аниме с сайта и возвращает []models.Anime,
// в которых обязательно заполнeны поля Title, EpCtx.AiredEpCount и EpCtx.TotalEpCount.
//
// SetAllEmbedLinks - подставляет все embed для аниме, если это возможно сделать не перегружая сервер.
// В противном случае получает дополнительные данные для заполнения embed ссылок поштучно.
// Если дополнительные данные не нужны, то не делает ничего. (TODO: придумать что-то красивее)
//
// SetEmbedLinks - устанавливает embed ссылки на эпизод ep из структуры anime.
// Если ссылки были установлены с помощью SetAllEmbedLinks, то не делает ничего.
//
// PrepareSavedAnime - дозаполняет структуру аниме из БД необходимыми данными для SetAllEmbedLinks и SetEmbedLinks.
type animeParser interface {
	GetAnimesByTitle(title string) ([]models.Anime, error)
	SetAllEmbedLinks(anime *models.Anime) error
	SetEmbedLinks(anime *models.Anime, ep *models.Episode) error
	PrepareSavedAnime(anime *models.Anime) error
}

func GetProvidersState(providers map[string]string) string {
	dialer := httpkit.NewDialer()
	var wg sync.WaitGroup
	var mu sync.Mutex

	var b strings.Builder
	for key, provider := range providers {
		key := key
		provider := provider
		providerLink := "https://" + provider

		wg.Add(1)
		go func() {
			defer wg.Done()

			if _, err := dialer.Dial(providerLink); err != nil {
				mu.Lock()
				defer mu.Unlock()
				fmt.Fprintf(&b, "Источник %s - %s не доступен\n", key, provider)
			} else {
				mu.Lock()
				defer mu.Unlock()
				fmt.Fprintf(&b, "Источник %s - %s доступен\n", key, provider)
			}
		}()
	}

	wg.Wait()

	return b.String()
}

func NewAPI(providerDomainMap map[string]string, playerDomains []string, dbh *db.DBHandler) (*API, error) {
	animeParsers := make(map[string]animeParser)

	for name, fullDomain := range providerDomainMap {
		animeParser, err := newAnimeParserByName(name, fullDomain)
		if err != nil {
			return nil, err
		}
		animeParsers[name] = animeParser
	}

	if len(animeParsers) == 0 {
		return nil, errors.New("не удалось найти ни один парсер в конфиге")
	}

	var converter *player.PlayerLinkConverter
	var err error
	converter, err = player.NewPlayerLinkConverter(playerDomains)
	if err != nil {
		return nil, err
	}

	a := API{
		animeParsers: animeParsers,
		Converter:    converter,
	}
	return &a, nil
}

func (a *API) GetAnimesByTitle(title string) ([]models.Anime, error) {
	var animes []models.Anime
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, client := range a.animeParsers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			parsedAnimes, err := client.GetAnimesByTitle(title)
			if err != nil {
				logger.ErrorLog.Println(err)
			}

			mu.Lock()
			defer mu.Unlock()
			animes = append(animes, parsedAnimes...)
		}()
	}
	wg.Wait()

	if len(animes) == 0 {
		return nil, errors.New("по вашему запросу ничего не найдено")
	}

	// Отбрасывание дубликатов
	dupRem := providers.NewDuplicateRemover()
	animes, err := dupRem.Remove(animes)
	if err != nil {
		return nil, err
	}
	sortBySearchPos(animes)

	return animes, nil
}

// Пытается установить все embed если это возможно.
func (a *API) SetAllEmbedLinks(anime *models.Anime) error {
	client := a.animeParsers[anime.Provider]
	err := client.SetAllEmbedLinks(anime)
	if err != nil {
		return err
	}

	return nil
}

func (a *API) SetEmbedLinks(anime *models.Anime, ep *models.Episode) error {
	client := a.animeParsers[anime.Provider]

	err := client.SetEmbedLinks(anime, ep)
	if err != nil {
		return err
	}

	return nil
}

func (a *API) PrepareSavedAnime(anime *models.Anime) error {
	client, ok := a.animeParsers[anime.Provider]
	if !ok {
		// Зануляем источник, если его больше нет в конфиге
		anime.Provider = ""
		return fmt.Errorf("парсер %s не доступен, проверьте конфиг", anime.Provider)
	}
	if err := client.PrepareSavedAnime(anime); err != nil {
		// Зануляем источник, если больше нет ответа
		anime.Provider = ""
		return err
	}
	return nil
}
