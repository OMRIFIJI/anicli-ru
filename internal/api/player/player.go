package player

import (
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/models"
	"anicliru/internal/api/player/aniboom"
)

type embedHandler interface {
	FindLinks(string) (map[string][]string, error)
}

func GetVideoLink(embedLink models.EmbedLink) (models.VideoLink, error) {
	handlers := getPlayerHandlers()
	videoLink := make(models.VideoLink)

	for dubName, dubPlayerLinks := range embedLink {
		videoLink[dubName] = make(map[string][]string)

		for playerName, link := range dubPlayerLinks {
			handler, exists := handlers[playerName]
			if !exists {
				apilog.WarnLog.Printf("Нет реализации обработки плеера %s", playerName)
				continue
			}

			qualityToLink, err := handler.FindLinks(link)
			if err != nil {
				continue
			}

			for quality := range qualityToLink {
				videoLink[dubName][quality] = append(videoLink[dubName][quality], qualityToLink[quality]...)
			}
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

func getPlayerHandlers() map[string]embedHandler {
	handlers := make(map[string]embedHandler)
	handlers["aniboom"] = aniboom.NewAniboom()

	return handlers
}
