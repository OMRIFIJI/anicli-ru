package kodik

import (
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/player/common"
	httpkit "github.com/OMRIFIJI/anicli-ru/internal/httpkit"
	"github.com/OMRIFIJI/anicli-ru/internal/logger"
	"bytes"
	"encoding/base64"
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
	Origin       = common.Kodik
	baseUrl      = "https://kodik.info"
	headerFields = `--http-header-fields="Referer: https://animego.org","Accept-Language: ru-RU"`
)

type Kodik struct {
	client     *httpkit.HttpClient
	postClient http.Client
	baseUrl    string
	apiPath    string
}

func NewKodik() *Kodik {
	client := httpkit.NewHttpClient(
		map[string]string{
			"Referer":         common.DefaultReferer,
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

	k := Kodik{
		client:     client,
		baseUrl:    baseUrl,
		apiPath:    baseUrl + "/ftor",
		postClient: postClient,
	}

	return &k
}

func (k *Kodik) GetVideos(embedLink string) (map[int]common.DecodedEmbed, error) {
	embedLink = "https:" + embedLink
	res, err := k.client.Get(embedLink)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	payload, err := k.getApiPayload(resBody)
	if err != nil {
		return nil, err
	}

	clientApi := httpkit.NewHttpClient(
		map[string]string{
			"Origin":  k.baseUrl,
			"Referer": embedLink,
			"Accept":  "application/json, text/javascript, */*; q=0.01",
		},
		httpkit.FromClient(&k.postClient),
	)

	resApi, err := clientApi.Post(k.apiPath, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resApi.Body.Close()

	resApiBody, err := io.ReadAll(resApi.Body)
	if err != nil {
		return nil, err
	}

	var vidData videoData
	err = json.Unmarshal(resApiBody, &vidData)
	if err != nil {
		return nil, err
	}

	links := k.videoDataToLinks(vidData)
	if len(links) == 0 {
		return nil, errors.New("не найдено ни одной ссылки")
	}

	return links, nil
}

func (k *Kodik) videoDataToLinks(vidData videoData) map[int]common.DecodedEmbed {
	links := make(map[int]common.DecodedEmbed)

	mpvOpts := []string{
		headerFields,
	}

	for key, val := range vidData.Links {
		if len(val) == 0 {
			logger.ErrorLog.Println("Ошибка обработки json в Kodik.")
			continue
		}
		decodedUrl, err := decodeUrl(val[0].Src)
		if err != nil {
			logger.ErrorLog.Printf("Ошибка расшифровки %s в Kodik. %s\n", val[0].Src, err)
			continue
		}
		ind := strings.Index(decodedUrl, ":hls:manifest")

		var link string
		if ind != -1 {
			link = decodedUrl[:ind]
		} else {
			link = decodedUrl
		}

		quality, err := strconv.Atoi(key)
		if err != nil {
			logger.ErrorLog.Printf("Ошибка обработки качества видео в Kodik. %s\n", err)
			continue
		}

		video := models.Video{
			Link:    link,
			MpvOpts: mpvOpts,
		}
		links[quality] = common.DecodedEmbed{
			Video:  video,
			Origin: Origin,
		}
	}

	return links
}

func (k *Kodik) getApiPayload(resBody []byte) ([]byte, error) {
	exp := `var domain = "(.+)";\s+var d_sign = "(.+)";\s+var pd = "(.+)";\s+var pd_sign = "(.+)";\s+var ref = "(.+)";\s+var ref_sign = "(.+)";`
	re := regexp.MustCompile(exp)
	match := re.FindStringSubmatch(string(resBody))
	if match == nil {
		return nil, errors.New("ответ от kodik не удалось обработать")
	}

	payloadMap := make(map[string]string)
	payloadMap["d"] = match[1]
	payloadMap["d_sign"] = match[2]
	payloadMap["pd"] = match[3]
	payloadMap["pd_sign"] = match[4]
	payloadMap["ref"] = match[5]
	payloadMap["ref_sign"] = match[6]

	exp = `videoInfo\.type\s+=\s+'(.+?)';\s+videoInfo\.hash\s+=\s+'(.+?)';\s+videoInfo\.id\s+=\s+'(.+?)'`
	re = regexp.MustCompile(exp)
	match = re.FindStringSubmatch(string(resBody))
	if match == nil {
		return nil, errors.New("ответ от kodik не удалось обработать")
	}
	payloadMap["type"] = match[1]
	payloadMap["hash"] = match[2]
	payloadMap["id"] = match[3]

	var b strings.Builder
	for key, val := range payloadMap {
		fmt.Fprintf(&b, "&%s=%s", key, val)
	}
	payload := []byte(b.String())

	return payload, nil
}

func decodeRot13(input string) string {
	var result strings.Builder
	for _, char := range input {
		switch {
		case 'a' <= char && char <= 'z':
			result.WriteRune((char-'a'+13)%26 + 'a')
		case 'A' <= char && char <= 'Z':
			result.WriteRune((char-'A'+13)%26 + 'A')
		default:
			result.WriteRune(char)
		}
	}
	return result.String()
}

func padBase64(base64Str string) string {
	padLength := len(base64Str) % 4
	if padLength != 0 {
		base64Str += strings.Repeat("=", 4-padLength)
	}
	return base64Str
}

func decodeUrl(urlEncoded string) (string, error) {
	if isDecoded(urlEncoded) {
		return common.AppendHttp(urlEncoded), nil
	}

	base64URL := decodeRot13(urlEncoded)

	base64URL = padBase64(base64URL)
	decodedBytes, err := base64.StdEncoding.DecodeString(base64URL)
	if err != nil {
		return "", err
	}
	decodedURL := string(decodedBytes)

	decodedURL = common.AppendHttp(decodedURL)

	return decodedURL, nil
}

func isDecoded(url string) bool {
	return strings.Contains(url, "cloud.kodik")
}
