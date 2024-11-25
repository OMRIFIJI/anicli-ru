package httpcommon

import (
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/models"
	"io"
	"net/http"
	"time"
)

type HttpClient struct {
	client  http.Client
	Headers map[string]string
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
		Headers: headers,
	}

	return &a
}

func (a *HttpClient) Get(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	for key, val := range a.Headers {
		req.Header.Add(key, val)
	}

	res, err := a.client.Do(req)
	if err != nil {
		apilog.ErrorLog.Printf("Http error. %s\n", err)
		return nil, err
	}

	if res.StatusCode != 200 {
		noConError := models.HttpError{
			Msg: "Не удалось соединиться с сервером. Код ошибки: " + res.Status,
		}
		res.Body.Close()
		return nil, &noConError
	}

	return res, err
}

func (a *HttpClient) Post(link string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", link, body)
	if err != nil {
		return nil, err
	}
	for key, val := range a.Headers {
		req.Header.Add(key, val)
	}

	res, err := a.client.Do(req)
	if err != nil {
		apilog.ErrorLog.Printf("Http error. %s\n", err)
		return nil, err
	}

	if res.StatusCode != 200 {
		noConError := models.HttpError{
			Msg: "Не удалось соединиться с сервером. Код ошибки: " + res.Status,
		}
		res.Body.Close()
		return nil, &noConError
	}

	return res, err
}
