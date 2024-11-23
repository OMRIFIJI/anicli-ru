package aniboom

import (
	apilog "anicliru/internal/api/log"
	httpcommon "anicliru/internal/http"
	"io"
)

type Aniboom struct {
	client *httpcommon.HttpClient
}

func NewAniboom() *Aniboom {
	client := httpcommon.NewHttpClient(
		map[string]string{
			"Referer":         "https://animego.one/",
			"Accept-Language": "ru-RU",
			"Origin":          "https://aniboom.one",
		},
	)

	a := Aniboom{
		client: client,
	}
	return &a
}

func (a *Aniboom) FindLinks(embedLink string) (map[string][]string, error) {
	embedLink = "https:" + embedLink
	res, err := a.client.Get(embedLink)
	if err != nil {
		apilog.ErrorLog.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	r, _ := io.ReadAll(res.Body)
	apilog.WarnLog.Println(string(r))

	return nil, nil
}
