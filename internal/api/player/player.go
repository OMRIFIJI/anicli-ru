package player

import (
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/models"
	"anicliru/internal/api/player/aniboom"
	"anicliru/internal/api/player/kodik"
	"anicliru/internal/api/player/sibnet"
	"sync"
)

type embedHandler interface {
	GetVideos(string) (map[int]models.Video, error)
}

type PlayerLinkConverter struct {
	handlers map[string]embedHandler
}

func (p *PlayerLinkConverter) SetPlayerHandlers() {
	handlers := make(map[string]embedHandler)
	handlers["aniboom.one"] = aniboom.NewAniboom()
	handlers["kodik.info"] = kodik.NewKodik()
	handlers["video.sibnet.ru"] = sibnet.NewSibnet()

	p.handlers = handlers
}

// Было бы неплохо заменить на каналы
func (p *PlayerLinkConverter) GetVideos(embedLinks models.EmbedLinks) (models.VideoLinks, error) {
	videoLinks := make(models.VideoLinks)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for dubName := range embedLinks {
		videoLinks[dubName] = make(map[int]models.Video)
	}

	for dubName, playerLinks := range embedLinks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for playerName, link := range playerLinks {
				handler, exists := p.handlers[playerName]
				if !exists {
					apilog.WarnLog.Printf("Нет реализации обработки плеера %s %s", playerName, link)
					return
				}

				qualityToVideo, err := handler.GetVideos(link)
				if err != nil {
					apilog.ErrorLog.Printf("Ошибка обработки плеера %s, %s", playerName, err)
					continue
				}

				mu.Lock()
				for quality := range qualityToVideo {
					_, exists := videoLinks[dubName][quality]
					if !exists {
						videoLinks[dubName][quality] = qualityToVideo[quality]
					}
				}
				mu.Unlock()
			}

		}()
	}
	wg.Wait()

	for dubName := range videoLinks {
		if len(videoLinks[dubName]) == 0 {
			delete(videoLinks, dubName)
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
