package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type AnilibriaAPI struct {
	Client       *http.Client
	BaseURL      string
	SearchMethod string
}

type AnimeSearchResponse struct {
	List []AnimeInfo `json:"list"`
}

type AnimeInfo struct {
	ID    int            `json:"id"`
	Names LanguageTitles `json:"names"`
	Media MediaInfo      `json:"player"`
}

type LanguageTitles struct {
	RU string `json:"ru"`
}

type MediaInfo struct {
	Host     string             `json:"host"`
	Episodes map[string]Episode `json:"List"`
}

type Episode struct {
	Number   int        `json:"episode"`
	HLSLinks HLSOptions `json:"hls"`
}

type HLSOptions struct {
	FHD string `json:"fhd"`
	HD  string `json:"hd"`
	SD  string `json:"sd"`
}

func (a *AnilibriaAPI) SearchTitleByName(titleName string) (*AnimeSearchResponse, error) {
	var searchRes AnimeSearchResponse

	searchRequest := fmt.Sprintf(
		"%s%s?search=%s&filter=id,names.ru,player.host,player.list&limit=50",
		a.BaseURL,
		a.SearchMethod,
		titleName,
	)

	res, err := a.Client.Get(searchRequest)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("Не получилось соединиться с сервером. Код ошибки: " + res.Status)
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&searchRes); err != nil {
		return nil, err
	}
	return &searchRes, nil
}

// Возвращает map элемент которого - это map из 3 ссылок разбитых по качеству
func (a *AnimeSearchResponse) GetLinks(indTitle int) map[string]map[string]string {
    episodesLinks := make(map[string]map[string]string)
	host := a.List[indTitle].Media.Host
    if !strings.HasPrefix(host, "http"){
        host = "https://" + host
    }
	episodes := a.List[indTitle].Media.Episodes
    for key, val := range episodes {
        linksMap := map[string]string{
            "FHD": host + val.HLSLinks.FHD,
            "HD": host + val.HLSLinks.HD,
            "SD": host + val.HLSLinks.SD,
        }
        episodesLinks[key] = linksMap
    }
    return episodesLinks
}
