package common

import (
	"net/http"
	"time"
)

func InitHttpClient() http.Client {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    60 * time.Second,
		DisableCompression: true,
	}
	client := http.Client{Transport: tr}
	return client
}
