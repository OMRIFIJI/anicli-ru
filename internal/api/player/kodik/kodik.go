package kodik

import (
	"anicliru/internal/logger"
	"anicliru/internal/api/models"
	"anicliru/internal/api/player/common"
	httpcommon "anicliru/internal/http"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const Netloc = "kodik.info"

type Kodik struct {
	client  *httpcommon.HttpClient
	baseUrl string
	name    string
}

func NewKodik() *Kodik {
	client := httpcommon.NewHttpClient(
		map[string]string{
			"Referer":         "https://animego.org/",
			"Accept-Language": "ru-RU",
		},
		httpcommon.WithRetries(2),
	)

	k := Kodik{
		client:  client,
		baseUrl: "https://kodik.info/",
	}
	return &k
}

type kodikVideoData struct {
	Links map[string][]struct {
		Src string `json:"src"`
	} `json:"links"`
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

	apiPath := k.baseUrl + "ftor"

	clientApi := httpcommon.NewHttpClient(
		map[string]string{
			"Origin":  k.baseUrl,
			"Referer": embedLink,
			"Accept":  "application/json, text/javascript, */*; q=0.01",
		},
	)

	resApi, err := clientApi.Post(apiPath, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resApi.Body.Close()

	resApiBody, err := io.ReadAll(resApi.Body)
	if err != nil {
		return nil, err
	}

	var videoData kodikVideoData
	err = json.Unmarshal(resApiBody, &videoData)
	if err != nil {
		return nil, err
	}

	links := k.videoDataToLinks(videoData)
	if len(links) == 0 {
		return nil, errors.New("Не найдено ни одной ссылки")
	}

	return links, nil
}

func (k *Kodik) videoDataToLinks(videoData kodikVideoData) map[int]common.DecodedEmbed {
	links := make(map[int]common.DecodedEmbed)

	mpvOpts := []string{
		`--http-header-fields="Referer: https://aniboom.one","Accept-Language: ru-RU"`,
	}

	for key, val := range videoData.Links {
		if len(val) == 0 {
			logger.ErrorLog.Println("Ошибка обработки json в Kodik.")
			continue
		}
		decodedUrl, err := decodeUrl(val[0].Src)
		if err != nil {
			logger.ErrorLog.Printf("Ошибка расшифровки в Kodik. %s\n", err)
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
			Origin: Netloc,
		}
	}

	return links
}

func (k *Kodik) getApiPayload(resBody []byte) ([]byte, error) {
	exp := `var domain = "(.+)";\s+var d_sign = "(.+)";\s+var pd = "(.+)";\s+var pd_sign = "(.+)";\s+var ref = "(.+)";\s+var ref_sign = "(.+)";`
	re := regexp.MustCompile(exp)
	match := re.FindStringSubmatch(string(resBody))
	if match == nil {
		return nil, errors.New("Ответ от kodik не удалось обработать")
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
		return nil, errors.New("Ответ от kodik не удалось обработать")
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
	base64URL := decodeRot13(urlEncoded)

	base64URL = padBase64(base64URL)
	decodedBytes, err := base64.StdEncoding.DecodeString(base64URL)
	if err != nil {
		return "", err
	}
	decodedURL := string(decodedBytes)

	// Проверяем, начинается ли строка с "https"
	if !strings.HasPrefix(decodedURL, "https") {
		decodedURL = "https:" + decodedURL
	}

	return decodedURL, nil
}

func removeSlashes(link string) string {
	link = strings.ReplaceAll(link, `\\\`, "")
	return link
}
