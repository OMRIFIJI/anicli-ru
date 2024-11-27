package sibnet

import (
	httpcommon "anicliru/internal/http"
	"io"
	"time"
)

type Sibnet struct {
    client  *httpcommon.HttpClient
}


func NewSibnet() *Sibnet{
    client := httpcommon.NewHttpClient(
		map[string]string{
			"Referer": "https://animego.org/",
            "Accept-Language": "ru-RU",
		},
	)
    client.Client.Timeout = 1 * time.Second // Sibnet часто может не отвечать
    s := &Sibnet{client: client}
	return s
}

func (s *Sibnet) FindLinks(embedLinks string) (map[int]string, error) {
	embedLinks = "https:" + embedLinks
	res, err := s.client.Get(embedLinks)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
    return nil, nil
}
