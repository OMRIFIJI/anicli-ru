package video

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/app/config"
	"github.com/OMRIFIJI/anicli-ru/internal/logger"
)

const (
	mpvRetries    = 5
	mpvNetTimeout = 3
)

type videoPlayer struct {
	Videos      map[string]map[int]models.Video
	cfg         *config.VideoCfg
	ResolvedDub string
}

type noDubError struct {
	Msg string
}

func (nde *noDubError) Error() string {
	return nde.Msg
}

func newVideoPlayer(cfg *config.VideoCfg) *videoPlayer {
	// Приводит название озвучки к нижнему регистру
	return &videoPlayer{
		cfg: &config.VideoCfg{
			Dub:     strings.ToLower(cfg.Dub),
			Quality: cfg.Quality,
		},
	}
}

func (vp *videoPlayer) SetVideos(newVideos map[string]map[int]models.Video) error {
	vp.Videos = newVideos

	// Если озвучку уже расшифровывали и не нашлась
	if _, ok := vp.Videos[vp.ResolvedDub]; vp.ResolvedDub != "" && !ok {
		err := &noDubError{
			Msg: "выбранная озвучка больше не доступна или нашлось несколько подходящих",
		}
		return err
	}

	// Если не расшифровывали -> расшфируем
	if vp.ResolvedDub == "" {
		key, ok := vp.findDubKey()
		if !ok {
			err := &noDubError{
				Msg: "выбранная озвучка больше не доступна или нашлось несколько подходящих",
			}
			return err
		}

		vp.ResolvedDub = key
	}

	qualities, _ := vp.GetQualities(vp.ResolvedDub)
	vp.cfg.Quality = vp.getClosestQuality(vp.cfg.Quality, qualities)

	return nil
}

// Возвращает ключ, соответствующий озвучке, и true в случае успеха.
// Если подходит несколько озвучек, возвращает ("", false).
func (vp *videoPlayer) findDubKey() (string, bool) {
	key := ""

	i := 0
	dubs := vp.GetDubs()

	for _, dub := range dubs {
		if strings.Contains(strings.ToLower(dub), vp.cfg.Dub) {
			i++
			key = dub
		}
	}

	if i > 1 || key == "" {
		return "", false
	}

	return key, true
}

func (vp *videoPlayer) GetDubs() []string {
	dubs := make([]string, 0, len(vp.Videos))
	for dub := range vp.Videos {
		dubs = append(dubs, dub)
	}

	sort.Strings(dubs)
	return dubs
}

func (vp *videoPlayer) GetQualities(dub string) ([]int, error) {
	if qualities, exists := vp.Videos[dub]; exists {
		qualityList := make([]int, 0, len(qualities))
		for quality := range qualities {
			qualityList = append(qualityList, quality)
		}

		sort.Sort(sort.Reverse(sort.IntSlice(qualityList)))
		return qualityList, nil
	}

	return nil, fmt.Errorf("озвучка '%s' не найдена", dub)
}

func (vp *videoPlayer) GetVideo() (*models.Video, error) {
	if vp.ResolvedDub == "" {
		return nil, errors.New("озвучка не выбрана")
	}

	if qualities, exists := vp.Videos[vp.ResolvedDub]; exists {
		if video, exists := qualities[vp.cfg.Quality]; exists {
			return &video, nil
		}
		return nil, fmt.Errorf("качество %d не существует для озвучки '%s'", vp.cfg.Quality, vp.ResolvedDub)
	}

	return nil, fmt.Errorf("озвучка '%s' не найдена", vp.ResolvedDub)
}

func (vp *videoPlayer) SelectDub(dub string) error {
	if _, exists := vp.Videos[dub]; !exists {
		return &noDubError{Msg: fmt.Sprintf("озвучка '%s' не найдена", dub)}
	}
	vp.ResolvedDub = dub

	qualities, err := vp.GetQualities(dub)
	if err != nil {
		return err
	}

	if vp.cfg.Quality == 0 {
		vp.cfg.Quality = qualities[0]
		return nil
	}

	vp.cfg.Quality = vp.getClosestQuality(vp.cfg.Quality, qualities)

	return nil
}

func (vp *videoPlayer) SelectQuality(quality int) error {
	if vp.ResolvedDub == "" {
		return fmt.Errorf("озвучка не выбрана")
	}
	qualities, _ := vp.GetQualities(vp.ResolvedDub)
	vp.cfg.Quality = vp.getClosestQuality(quality, qualities)
	return nil
}

func (vp *videoPlayer) getClosestQuality(target int, qualities []int) int {
	if len(qualities) == 0 {
		return 0
	}

	closest := qualities[0]
	minDiff := intAbs(target - closest)

	for _, q := range qualities {
		diff := intAbs(target - q)
		if diff < minDiff {
			closest = q
			minDiff = diff
		}

		if minDiff == 0 {
			return closest
		}
	}

	return closest
}

func (vp *videoPlayer) StartMpv(title string, ctx context.Context) error {
	video, err := vp.GetVideo()
	if err != nil {
		return err
	}

	mpvOpts := video.MpvOpts
	mpvOpts = append(mpvOpts, fmt.Sprintf(`--force-media-title="%s"`, title))
	mpvOpts = append(mpvOpts, fmt.Sprintf("--network-timeout=%d", mpvNetTimeout))
	mpvOpts = append(mpvOpts, fmt.Sprintf(`"%s"`, video.Link))

	// Пока лучший способ, который нашёл, чтобы пережить обработку bash array для хэдеров
	mpvCmd := "mpv " + strings.Join(mpvOpts, " ")

	// Несколько раз пытаюсь достучаться до видео, особенно актуально для sibnet
	for i := range mpvRetries {
		cmd := execCommand(mpvCmd)

		if err := cmd.Start(); err != nil {
			logger.ErrorLog.Printf("не удалось запустить mpv на %d попытке. %s\n", i+1, err)
			continue
		}

		if cmd.Process == nil {
			logger.ErrorLog.Printf("не удача mpv на %d попытке: process == nil.\n", i+1)
			continue
		}

		if err := cmd.Wait(); err != nil {
			logger.ErrorLog.Printf("не удача mpv на %d попытке %s.\n", i+1, err)
			continue
		}

		return nil
	}

	return fmt.Errorf("не удалось запустить MPV после %d попыток", mpvRetries)
}

func intAbs(value int) int {
	if value < 0 {
		return (-1) * value
	}
	return value
}
