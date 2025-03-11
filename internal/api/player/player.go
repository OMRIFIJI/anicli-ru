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
	"anicliru/internal/logger"
	"errors"
	"sync"
)

type embedHandler interface {
	GetVideos(string) (map[int]common.DecodedEmbed, error)
}

type PlayerLinkConverter struct {
	handlers    map[string]embedHandler
	priorityMap map[string]int
}

func NewPlayerLinkConverter() *PlayerLinkConverter {
	plc := PlayerLinkConverter{}
	plc.SetPlayerHandlers()
	plc.SetPriorityMap()
	return &plc
}

func (plc *PlayerLinkConverter) SetPlayerHandlers() {
	handlers := make(map[string]embedHandler)
	handlers[aniboom.Netloc] = aniboom.NewAniboom()
	handlers[kodik.Netloc] = kodik.NewKodik()
	handlers[sibnet.Netloc] = sibnet.NewSibnet()
	handlers[vk.Netloc] = vk.NewVK()
	handlers[alloha.Netloc] = alloha.NewAlloha()
	handlers[aksor.Netloc] = aksor.NewAksor()
	handlers[sovrom.Netloc] = sovrom.NewSovrom()

	plc.handlers = handlers
}

// Задаёт приоритет плееров.
// При удалении дубликатов остаются видео плеера высшего приоритета.
func (plc *PlayerLinkConverter) SetPriorityMap() {
	plc.priorityMap = map[string]int{
		aniboom.Netloc: 6, // Высокий приоритет
		kodik.Netloc:   5,
		vk.Netloc:      4,
		alloha.Netloc:  3,
		aksor.Netloc:   2,
		sovrom.Netloc:  1,
		sibnet.Netloc:  0, // Низкий приоритет
	}
}

type workerDecodeRes struct {
	dubName  string
	dubLinks map[int][]common.DecodedEmbed
}

// Перемудрил
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
		handler, exists := plc.handlers[playerName]
		if !exists {
			logger.WarnLog.Printf("Нет реализации обработки плеера %s %s\n", playerName, link)
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
