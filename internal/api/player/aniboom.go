package player

import "net/http"

type AniBoom struct {
    headers map[string]string
    client http.Client
}

func NewAniBoom(client http.Client) *AniBoom {
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

func GetAniboomLinks(embedLink string) map[string]string{
    return nil
}
