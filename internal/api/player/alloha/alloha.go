package alloha

import (
	"github.com/OMRIFIJI/anicli-ru/internal/api/models"
	"github.com/OMRIFIJI/anicli-ru/internal/api/player/common"
	httpkit "github.com/OMRIFIJI/anicli-ru/internal/httpkit"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	Origin       = common.Alloha
	headerFields = `--http-header-fields="Origin: https://alloha.yani.tv/","Referer: https://animego.org/"`
    baseUrl = "https://alloha.yani.tv"
)

type Alloha struct {
	client     *httpkit.HttpClient
	postClient http.Client
}

func NewAlloha() *Alloha {
	client := httpkit.NewHttpClient(
		map[string]string{
			"Referer":         common.DefaultReferer,
			"Origin":          baseUrl,
			"Accept-Language": "ru-RU",
		},
		httpkit.WithRetries(2),
	)

	tr := &http.Transport{
		MaxIdleConns:       70,
		DisableCompression: true,
	}
	postClient := http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}

	a := Alloha{
		client:     client,
		postClient: postClient,
	}
	return &a
}

func (a *Alloha) GetVideos(embedLink string) (map[int]common.DecodedEmbed, error) {
	embedLink = common.AppendHttp(embedLink)
	animeId, acceptsContorls, payload, err := a.getPayload(embedLink)
	if err != nil {
		return nil, err
	}

	clientApi := httpkit.NewHttpClient(
		map[string]string{
			"Origin":           a.client.Headers["Origin"],
			"Referer":          embedLink,
			"Accepts-Controls": acceptsContorls,
			"Content-Type":     "application/x-www-form-urlencoded; charset=UTF-8",
			"Content-Length":   strconv.Itoa(len(payload)),
		},
		httpkit.FromClient(&a.postClient),
	)

	dest := fmt.Sprintf("%s/movie/%d", baseUrl, animeId)
	res, err := clientApi.Post(dest, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var vidData videoData

	if err := json.Unmarshal(resBody, &vidData); err != nil {
		return nil, err
	}

	qualityToLinks := vidData.HLSSources[0].Quality

	delete(qualityToLinks, "Object")

	links := make(map[int]common.DecodedEmbed)
	for qualityStr, link := range qualityToLinks {
		quality, err := strconv.Atoi(qualityStr)
		if err != nil {
			continue
		}

		// Может возвращать несколько ссылок, разделенных " or "
		endOfLinkInd := strings.Index(link, " or ")
		if endOfLinkInd != -1 {
			link = link[:endOfLinkInd]
		}

		links[quality] = common.DecodedEmbed{
			Video: models.Video{
				Link:    link,
				MpvOpts: []string{headerFields},
			},
			Origin: Origin,
		}
	}

	if len(links) == 0 {
		return nil, errors.New("не найдено ни одной ссылки")
	}

	return links, nil
}

func (a *Alloha) getPayload(embedLink string) (int, string, []byte, error) {
	res, err := a.client.Get(embedLink)
	if err != nil {
		return 0, "", nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, "", nil, err
	}

	re := regexp.MustCompile(`const\s+fileList\s+= JSON\.parse\('{"type":\s*"serial",\s*"active":\s*{"id":\s*(\d+)`)
	match := re.FindStringSubmatch(string(resBody))

	if match == nil {
		return 0, "", nil, errors.New("не удалось обработать ссылку на видео")
	}

	animeId, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, "", nil, fmt.Errorf("не удалось обработать api payload %s", err)
	}

	re = regexp.MustCompile(`const\s+userParam\s*=\s*{\s*\n*\s*token:\s*'(.+?)'`)
	match = re.FindStringSubmatch(string(resBody))

	if match == nil {
		return 0, "", nil, errors.New("не удалось обработать token")
	}
	token := match[1]

	re = regexp.MustCompile(`<meta\s+name\s*=\s*"user"\s+content\s*=\s*"(.+?)">`)
	match = re.FindStringSubmatch(string(resBody))

	if match == nil {
		return 0, "", nil, errors.New("не удалось обработать Accepts-Controls")
	}
	acceptsContorls := match[1]

	payload := []byte(fmt.Sprintf("token=%s&av1=true&autoplay=0&audio=&subtitle=", token))
	return animeId, acceptsContorls, payload, err
}
