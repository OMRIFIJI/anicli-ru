package aniboom

import (
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/models"
	httpcommon "anicliru/internal/http"
	"io"
)

func GetLinks(embedLink string) (map[string]string, error) {
    embedLink = "https:" + embedLink
	client := httpcommon.NewHttpClient(
		map[string]string{
			"Referer":         "https://animego.one/",
			"Accept-Language": "ru-RU",
			"Origin":          "https://aniboom.one",
		},
	)
    res, err := client.Get(embedLink)
    if err != nil {
        apilog.ErrorLog.Println(err)
        return nil, err
    }
    defer res.Body.Close()

    r, _ := io.ReadAll(res.Body)
    apilog.WarnLog.Println(string(r))

	return nil, nil
}


func fillEpLinks(embedLink models.EmbedLink) error {
	epLinks := make(models.EmbedLink)

	for dubName, dubPlayerLinks := range episode.EmbedLink {
		epLinks[dubName] = make(map[string]map[string]string)

		for playerName, embedLink := range dubPlayerLinks {
			switch playerName {
			case "aniboom":
                links, err := aniboom.GetLinks(embedLink)
                if err != nil {
                    continue
                }
                epLinks[dubName][playerName] = links
			}
		}

	}

	if len(epLinks) == 0 {
		err := &models.NotFoundError{
			Msg: "Не удалось найти эту серию.",
		}
		return err
	}

	episode.EpLink = epLinks
	return nil
}
