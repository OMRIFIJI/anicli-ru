package aniboom

import "net/http"

type AniBoom struct {
    headers map[string]string
    client http.Client
}

func newAniBoom(client http.Client) *AniBoom {
    a := AniBoom{
        client: client,
        headers: map[string]string{
            "Referer": "https://aniboom.one/",
            "Accept-Language": "ru-RU",
            "Origin": "https://aniboom.one",
        },
    }
    return &a
}

func GetLinks(embedLink string) map[string]string{
    return nil
}
