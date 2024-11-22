package httpcommon

import (
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/models"
	"net/http"
	"time"
)

type HttpClient struct {
	client  http.Client
	headers map[string]string
}

func NewHttpClient(headers map[string]string) *HttpClient {
	tr := &http.Transport{
		MaxIdleConns:       30,
		IdleConnTimeout:    60 * time.Second,
		DisableCompression: true,
	}
	client := http.Client{Transport: tr}

	a := HttpClient{
		client:  client,
		headers: headers,
	}

	return &a
}

func (a *HttpClient) Get(link string) (*http.Response, error) {
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
		noConError := models.HttpError{
			Msg: "Не получилось соединиться с сервером. Код ошибки: " + res.Status,
		}
		res.Body.Close()
		return nil, &noConError
	}

	return res, err
}
