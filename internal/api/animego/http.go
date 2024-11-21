package animego

import (
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/types"
	"net/http"
	"time"
)

type animeGoHttp struct {
	client  http.Client
	headers map[string]string
}

func newAnimeGoHttp() *animeGoHttp {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    60 * time.Second,
		DisableCompression: true,
	}
	client := http.Client{Transport: tr}

	a := animeGoHttp{
		client: client,
		headers: map[string]string{
			"User-Agent":       "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0",
			"X-Requested-With": "XMLHttpRequest",
		},
	}

    return &a
}

func (a *animeGoHttp) get(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	for key, val := range a.headers {
		req.Header.Add(key, val)
	}

	res, err := a.client.Do(req)
	if err != nil {
		apilog.ErrorLog.Printf("Http error. %s\n", err)
		return nil, err
	}

	if res.StatusCode != 200 {
		noConError := types.HttpError{
			Msg: "Не получилось соединиться с сервером. Код ошибки: " + res.Status,
		}
        res.Body.Close()
		return nil, &noConError
	}

	return res, err
}
