package httpcommon

import (
	"anicliru/internal/logger"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	timeoutClientDefault = 2
	maxRetriesDefault    = 1
	RetryDelayDefault    = 0
	maxIdleConnsDefault  = 20

	timeoutDialerDefault = 2
)

type HttpClient struct {
	Client     *http.Client
	Headers    map[string]string
	MaxRetries int
	RetryDelay time.Duration
	Timeout    time.Duration
}

type Dialer struct {
	client *http.Client
}

func NewHttpClient(headers map[string]string, options ...func(*HttpClient)) *HttpClient {
	hc := &HttpClient{
		Client:  nil,
		Headers: headers,
	}
	hc.MaxRetries = maxRetriesDefault
	hc.RetryDelay = RetryDelayDefault
	hc.Timeout = time.Duration(timeoutClientDefault) * time.Second

	for _, o := range options {
		o(hc)
	}

	if hc.Client == nil {
		tr := &http.Transport{
			MaxIdleConns:       maxIdleConnsDefault,
			DisableCompression: true,
		}
		client := http.Client{
			Transport: tr,
		}
		hc.Client = &client
	}

	hc.Client.Timeout = hc.Timeout

	return hc
}

func WithRetries(maxRetries int) func(*HttpClient) {
	return func(hc *HttpClient) {
		hc.MaxRetries = maxRetries
	}
}

func WithRetryDelay(delay int) func(*HttpClient) {
	return func(hc *HttpClient) {
		hc.RetryDelay = time.Duration(delay)
	}
}

func WithTimeout(timeInSeconds uint) func(*HttpClient) {
	return func(hc *HttpClient) {
		hc.Timeout = time.Duration(timeInSeconds) * time.Second
	}
}

func FromClient(client *http.Client) func(*HttpClient) {
	return func(hc *HttpClient) {
		hc.Client = client
	}
}

func (hc *HttpClient) delay() {
	if hc.RetryDelay != 0 {
		time.Sleep(hc.RetryDelay * time.Second)
	}
}

func (hc *HttpClient) Get(link string) (*http.Response, error) {
	var err error
	for i := 0; i < hc.MaxRetries; i++ {
		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			hc.delay()
			continue
		}
		for key, val := range hc.Headers {
			req.Header.Add(key, val)
		}

		res, err := hc.Client.Do(req)
		if err != nil {
			logger.WarnLog.Printf("Http error. %s\n", err)
			hc.delay()
			continue
		}

		if res.StatusCode != 200 {
			res.Body.Close()
			hc.delay()
			continue
		}

		return res, nil
	}

	return nil, fmt.Errorf("ошибка http после %d попыток. Ссылка: %s. Последняя ошибка: %s", hc.MaxRetries, link, err)
}

func (hc *HttpClient) Post(link string, body io.Reader) (*http.Response, error) {
	var err error
	for i := 0; i < hc.MaxRetries; i++ {
		req, err := http.NewRequest("POST", link, body)
		if err != nil {
			hc.delay()
			continue
		}
		for key, val := range hc.Headers {
			req.Header.Add(key, val)
		}

		res, err := hc.Client.Do(req)
		if err != nil {
			logger.WarnLog.Printf("Http error. %s\n", err)
			hc.delay()
			continue
		}

		if res.StatusCode != 200 {
			res.Body.Close()
			hc.delay()
			continue
		}

		return res, nil
	}

	return nil, fmt.Errorf("ошибка http после %d попыток. Ссылка: %s. Последняя ошибка: %s", hc.MaxRetries, link, err)
}

func NewDialer() *Dialer {
	timeout := time.Duration(timeoutDialerDefault) * time.Second
	return &Dialer{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (d *Dialer) Ping(url string) (int, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}
