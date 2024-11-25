package player

import (
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/models"
	"anicliru/internal/api/player/aniboom"
	"anicliru/internal/api/player/kodik"
	"sync"
)

type embedHandler interface {
	FindLinks(string) (map[int]string, error)
}

type PlayerLinkConverter struct {
	handlers map[string]embedHandler
}
func (p *PlayerLinkConverter) SetPlayerHandlers() {
	handlers := make(map[string]embedHandler)
	handlers["aniboom"] = aniboom.NewAniboom()
	handlers["kodik"] = kodik.NewKodik()

	p.handlers = handlers
}

func (p *PlayerLinkConverter) GetVideoLink(embedLink models.EmbedLink) (models.VideoLink, error) {
	videoLink := make(models.VideoLink)

	var wg sync.WaitGroup
    var mu sync.Mutex

	for dubName, dubPlayerLinks := range embedLink {
        wg.Add(1)
		go func() {
			defer wg.Done()

            mu.Lock()
            videoLink[dubName] = make(map[int]string)
            mu.Unlock()

			for playerName, link := range dubPlayerLinks {
				handler, exists := p.handlers[playerName]
				if !exists {
					apilog.WarnLog.Printf("Нет реализации обработки плеера %s", playerName)
					return
				}

				qualityToLink, err := handler.FindLinks(link)
				if err != nil {
					apilog.ErrorLog.Printf("Ошибка обработки плеера %s, %s", playerName, err)
                    continue
				}

                mu.Lock()
				for quality := range qualityToLink {
                    _, exists := videoLink[dubName][quality]
                    if !exists {
                        videoLink[dubName][quality] = qualityToLink[quality]
                    }
				}
                mu.Unlock()
			}

		}()
	}
    wg.Wait()

    for dubName := range videoLink {
        if len(videoLink[dubName]) == 0 {
            delete(videoLink, dubName)
        }
    }

	if len(videoLink) == 0 {
		err := &models.NotFoundError{
			Msg: "Не удалось найти эту серию.",
		}
		return nil, err
	}

	return videoLink, nil
}
