package player

import (
	"anicliru/internal/api/models"
	"anicliru/internal/api/player/aksor"
	"anicliru/internal/api/player/alloha"
	"anicliru/internal/api/player/aniboom"
	"anicliru/internal/api/player/common"
	"anicliru/internal/api/player/kodik"
	"anicliru/internal/api/player/sibnet"
	"anicliru/internal/api/player/sovrom"
	"anicliru/internal/api/player/vk"
	config "anicliru/internal/app/cfg"
	"anicliru/internal/db"
	httpcommon "anicliru/internal/http"
	"anicliru/internal/logger"
	"errors"
	"sync"
)

type embedHandler interface {
	GetVideos(string) (map[int]common.DecodedEmbed, error)
}

type PlayerLinkConverter struct {
	Handlers        map[common.PlayerOrigin]embedHandler
	priorityMap     map[common.PlayerOrigin]int
	playerOriginMap map[string]common.PlayerOrigin
}

func NewPlayerLinkConverter(opts ...func(*PlayerLinkConverter, map[string]func() embedHandler)) (*PlayerLinkConverter, error) {
	playerOriginMap := common.NewPlayerOriginMap()
	priorityMap := getPriorityMap()

	plc := PlayerLinkConverter{
		priorityMap:     priorityMap,
		playerOriginMap: playerOriginMap,
	}

	newHandlerMap := map[string]func() embedHandler{
		common.AniboomDomain: func() embedHandler { return aniboom.NewAniboom() },
		common.KodikDomain:   func() embedHandler { return kodik.NewKodik() },
		common.SibnetDomain:  func() embedHandler { return sibnet.NewSibnet() },
		common.VKDomain:      func() embedHandler { return vk.NewVK() },
		common.AllohaDomain:  func() embedHandler { return alloha.NewAlloha() },
		common.AksorDomain:   func() embedHandler { return aksor.NewAksor() },
		common.SovromDomain:  func() embedHandler { return sovrom.NewSovrom() },
	}

	for _, o := range opts {
		o(&plc, newHandlerMap)
	}

	return &plc, nil
}

func FromConfig(cfg *config.Config) func(*PlayerLinkConverter, map[string]func() embedHandler) {
	return func(plc *PlayerLinkConverter, newHandlerMap map[string]func() embedHandler) {
		handlers := make(map[common.PlayerOrigin]embedHandler)

		for _, domain := range cfg.Players.Domains {
			origin := plc.playerOriginMap[domain]
			handlers[origin] = newHandlerMap[domain]()
		}

		plc.Handlers = handlers
	}
}

func WithSync(db *db.DBHandler) func(*PlayerLinkConverter, map[string]func() embedHandler) {
	return func(plc *PlayerLinkConverter, newHandlerMap map[string]func() embedHandler) {
		handlers := make(map[common.PlayerOrigin]embedHandler)

		dialer := httpcommon.NewDialer()
		var wg sync.WaitGroup

		domainChan := make(chan string)

		// Добавляет handler плеера, если сервер плеера отвечает
		addHandler := func(domain string) {
			wg.Add(1)
			go func() {
				defer wg.Done()

				url := "https://" + domain
				if _, err := dialer.Ping(url); err != nil {
					return
				}

				domainChan <- domain
			}()
		}

		for key, _ := range plc.playerOriginMap {
			addHandler(key)
		}

		go func() {
			wg.Wait()
			close(domainChan)
		}()

		var reachableDomains []string
		for domain := range domainChan {
			reachableDomains = append(reachableDomains, domain)

			origin := plc.playerOriginMap[domain]
			handlers[origin] = newHandlerMap[domain]()
		}

		plc.Handlers = handlers
	}
}

// Задаёт приоритет плееров.
// При удалении дубликатов остаются видео плеера высшего приоритета.
func getPriorityMap() map[common.PlayerOrigin]int {
	return map[common.PlayerOrigin]int{
		aniboom.Origin: 6, // Высокий приоритет
		kodik.Origin:   5,
		vk.Origin:      4,
		alloha.Origin:  3,
		aksor.Origin:   2,
		sovrom.Origin:  1,
		sibnet.Origin:  0, // Низкий приоритет
	}
}

type workerDecodeRes struct {
	dubName  string
	dubLinks map[int][]common.DecodedEmbed
}

func (plc *PlayerLinkConverter) GetVideos(embedLinks models.EmbedLinks) (models.VideoLinks, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	videoLinks := make(models.VideoLinks)
	for dubName, playerLinks := range embedLinks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			plc.decodeDub(dubName, playerLinks, videoLinks, &mu)
		}()
	}
	wg.Wait()

	if len(videoLinks) == 0 {
		return nil, errors.New("не удалось найти эту серию")
	}

	return videoLinks, nil
}

func (plc *PlayerLinkConverter) decodeDub(dubName string, playerLinks map[string]string, videoLinks models.VideoLinks, mu *sync.Mutex) {
	dubLinks := make(map[int][]common.DecodedEmbed)
	for playerName, link := range playerLinks {
		playerOrigin, ok := plc.playerOriginMap[playerName]
		if !ok {
			logger.WarnLog.Printf("Нет реализации обработки плеера %s %s\n", playerName, link)
			return
		}
		handler, ok := plc.Handlers[playerOrigin]
		if !ok {
			return
		}

		qualityToVideo, err := handler.GetVideos(link)
		if err != nil {
			logger.ErrorLog.Printf("Ошибка обработки плеера %s, %s\n", playerName, err)
			continue
		}

		for quality := range qualityToVideo {
			dubLinks[quality] = append(dubLinks[quality], qualityToVideo[quality])
		}
	}

	if len(dubLinks) == 0 {
		return
	}

	dubRes := workerDecodeRes{
		dubName:  dubName,
		dubLinks: dubLinks,
	}

	mu.Lock()
	defer mu.Unlock()
	videoLinks[dubRes.dubName] = make(map[int]models.Video)
	for quality, decodedEmbed := range dubRes.dubLinks {
		videoLinks[dubRes.dubName][quality] = plc.bestVideo(decodedEmbed)
	}
}

func (plc *PlayerLinkConverter) isOriginGreater(a, b common.DecodedEmbed) bool {
	return plc.priorityMap[a.Origin] > plc.priorityMap[b.Origin]
}

func (plc *PlayerLinkConverter) bestVideo(decodedEmbed []common.DecodedEmbed) models.Video {
	bestDecode := decodedEmbed[0]

	for _, decode := range decodedEmbed {
		if plc.isOriginGreater(decode, bestDecode) {
			bestDecode = decode
		}
	}

	return bestDecode.Video
}
