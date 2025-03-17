package player

import (
	"errors"
	"sync"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/aksor"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/alloha"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/aniboom"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/common"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/kodik"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/sibnet"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/sovrom"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/vk"
	httpkit "github.com/OMRIFIJI/anicli-ru/internal/httpkit"
	"github.com/OMRIFIJI/anicli-ru/internal/logger"
)

type embedHandler interface {
	GetVideos(string) (map[int]common.DecodedEmbed, error)
}

type PlayerLinkConverter struct {
	Handlers        map[common.PlayerOrigin]embedHandler
	priorityMap     map[common.PlayerOrigin]int
	playerOriginMap map[string]common.PlayerOrigin
}

func NewPlayerLinkConverter(playerDomains []string) (*PlayerLinkConverter, error) {
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

	handlers := make(map[common.PlayerOrigin]embedHandler)

	for _, domain := range playerDomains {
		origin := plc.playerOriginMap[domain]
		handlers[origin] = newHandlerMap[domain]()
	}

	plc.Handlers = handlers

	return &plc, nil
}

func SyncedDomains() []string {
	dialer := httpkit.NewDialer()
	playerOriginMap := common.NewPlayerOriginMap()

	var wg sync.WaitGroup
	var mu sync.Mutex

	var reachableDomains []string
	for domain := range playerOriginMap {
		wg.Add(1)
		go func() {
			defer wg.Done()

			url := "https://" + domain
			if _, err := dialer.Dial(url); err != nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()
			reachableDomains = append(reachableDomains, domain)
		}()
	}

	wg.Wait()
	return reachableDomains
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

type decodedDub struct {
	dubName string
	links   map[int][]common.DecodedEmbed
}

func (plc *PlayerLinkConverter) Convert(embedLinks models.EmbedLinks) (models.VideoLinks, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	videoLinks := make(models.VideoLinks)
	for dubName, playerLinks := range embedLinks {
		wg.Add(1)
		go func() {
			defer wg.Done()

			decDub := plc.convertDub(dubName, playerLinks)
			if decDub == nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()
			videoLinks[decDub.dubName] = make(map[int]models.Video)
			for quality, decodedEmbed := range decDub.links {
				videoLinks[decDub.dubName][quality] = plc.bestVideo(decodedEmbed)
			}
		}()
	}
	wg.Wait()

	if len(videoLinks) == 0 {
		return nil, errors.New("не удалось найти эту серию")
	}

	return videoLinks, nil
}

func (plc *PlayerLinkConverter) convertDub(dubName string, playerLinks map[string]string) *decodedDub {
	links := make(map[int][]common.DecodedEmbed)
	for playerName, link := range playerLinks {
		playerOrigin, ok := plc.playerOriginMap[playerName]
		if !ok {
			logger.WarnLog.Printf("Нет реализации обработки плеера %s %s\n", playerName, link)
			return nil
		}
		handler, ok := plc.Handlers[playerOrigin]
		if !ok {
			return nil
		}

		qualityToVideo, err := handler.GetVideos(link)
		if err != nil {
			logger.ErrorLog.Printf("Ошибка обработки плеера %s, %s\n", playerName, err)
			continue
		}

		for quality := range qualityToVideo {
			links[quality] = append(links[quality], qualityToVideo[quality])
		}
	}

	if len(links) == 0 {
		return nil
	}

	return &decodedDub{
		dubName: dubName,
		links:   links,
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
