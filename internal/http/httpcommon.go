package httpcommon

import (
	apilog "anicliru/internal/api/log"
	"anicliru/internal/api/models"
	"io"
	"net/http"
	"time"
)

type HttpClient struct {
	Client  http.Client
	Headers map[string]string
}

func NewHttpClient(headers map[string]string) *HttpClient {
	tr := &http.Transport{
		MaxIdleConns:       30,
		DisableCompression: true,
	}
	Client := http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	hc := HttpClient{
		Client:  Client,
		Headers: headers,
	}

	return &hc
}

func (hc *HttpClient) Get(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	for key, val := range hc.Headers {
		req.Header.Add(key, val)
	}

	res, err := hc.Client.Do(req)
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

func (hc *HttpClient) Post(link string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", link, body)
	if err != nil {
		return nil, err
	}
	for key, val := range hc.Headers {
		req.Header.Add(key, val)
	}

	res, err := hc.Client.Do(req)
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
