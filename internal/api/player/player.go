package player

import (
	"anicliru/internal/api/models"
	"anicliru/internal/api/player/aniboom"
	"anicliru/internal/api/player/common"
	"anicliru/internal/api/player/kodik"
	"anicliru/internal/api/player/sibnet"
	"anicliru/internal/api/player/vk"
	"sync"
)

type embedHandler interface {
	GetVideos(string) (map[int]common.DecodedEmbed, error)
}

type PlayerLinkConverter struct {
	handlers map[string]embedHandler
}

func (plc *PlayerLinkConverter) SetPlayerHandlers() {
	handlers := make(map[string]embedHandler)
	handlers[aniboom.Netloc] = aniboom.NewAniboom()
	handlers[kodik.Netloc] = kodik.NewKodik()
	handlers[sibnet.Netloc] = sibnet.NewSibnet()
	handlers[vk.Netloc] = vk.NewVK()

	plc.handlers = handlers
}

type workerDecodeRes struct {
	dubName  string
	dubLinks map[int][]common.DecodedEmbed
}

// Перемудрил
func (plc *PlayerLinkConverter) GetVideos(embedLinks models.EmbedLinks) (models.VideoLinks, error) {
	var wg sync.WaitGroup
	dubDecodeChan := make(chan workerDecodeRes)

	for dubName, playerLinks := range embedLinks {
		wg.Add(1)
		go func() {
			defer wg.Done()
            plc.decodeDub(dubName, playerLinks, dubDecodeChan)
		}()
	}

	go func() {
		defer close(dubDecodeChan)
		wg.Wait()
	}()

	videoLinks := make(models.VideoLinks)
	for dubRes := range dubDecodeChan {
		videoLinks[dubRes.dubName] = make(map[int]models.Video)
		for quality, decodedEmbed := range dubRes.dubLinks {
			videoLinks[dubRes.dubName][quality] = bestVideo(decodedEmbed)
		}
	}

	if len(videoLinks) == 0 {
		err := &models.NotFoundError{
			Msg: "Не удалось найти эту серию.",
		}
		return nil, err
	}

	return videoLinks, nil
}

func (plc *PlayerLinkConverter) decodeDub(dubName string, playerLinks map[string]string, dubDecodeChan chan workerDecodeRes) {
	dubLinks := make(map[int][]common.DecodedEmbed)
	for playerName, link := range playerLinks {
		handler, exists := plc.handlers[playerName]
		if !exists {
			return
		}

		qualityToVideo, err := handler.GetVideos(link)
		if err != nil {
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
	dubDecodeChan <- dubRes
}

func IsOriginGreater(a, b common.DecodedEmbed) bool {
	switch a.Origin {
	case aniboom.Netloc:
		return true
	case kodik.Netloc:
		if b.Origin == aniboom.Netloc {
			return false
		}
		return true
    case vk.Netloc:
        switch b.Origin{
        case aniboom.Netloc, kodik.Netloc:
            return false
        }
        return true
	case sibnet.Netloc:
		return false
	}

	return false
}

func bestVideo(decodedEmbed []common.DecodedEmbed) models.Video {
	bestDecode := decodedEmbed[0]

	for _, decode := range decodedEmbed {
		if IsOriginGreater(decode, bestDecode) {
			bestDecode = decode
		}
	}

	return bestDecode.Video
}
