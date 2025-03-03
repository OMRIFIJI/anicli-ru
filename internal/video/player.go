package video

import (
	"anicliru/internal/api/models"
	"anicliru/internal/logger"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	mpvRetries    = 5
	mpvNetTimeout = 3
)

type videoPlayer struct {
	Videos map[string]map[int]models.Video
	cfg    videoPlayerConfig
}

type noDubError struct {
	Msg string
}

func (nde *noDubError) Error() string {
	return nde.Msg
}

type videoPlayerConfig struct {
	CurrentDub     string
	CurrentQuality int
}

func (vpc *videoPlayerConfig) isEmpty() bool {
	return vpc.CurrentQuality == 0 && vpc.CurrentDub == ""
}

func newVideoPlayer() *videoPlayer {
	return &videoPlayer{
		cfg: videoPlayerConfig{
			CurrentQuality: 0,
			CurrentDub:     "",
		},
	}
}

func (vp *videoPlayer) SetVideos(newVideos map[string]map[int]models.Video) error {
	vp.Videos = newVideos

	if _, exists := vp.Videos[vp.cfg.CurrentDub]; !exists {
		vp.cfg.CurrentDub = ""
		err := &noDubError{
			Msg: "Выбранная озвучка больше не доступна.",
		}
		return err
	}

	qualities, _ := vp.GetQualities(vp.cfg.CurrentDub)
	vp.cfg.CurrentQuality = vp.getClosestQuality(vp.cfg.CurrentQuality, qualities)

	return nil
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
	if vp.cfg.CurrentDub == "" {
		return nil, errors.New("озвучка не выбрана")
	}

	if qualities, exists := vp.Videos[vp.cfg.CurrentDub]; exists {
		if video, exists := qualities[vp.cfg.CurrentQuality]; exists {
			return &video, nil
		}
		return nil, fmt.Errorf("качество %d не существует для озвучки '%s'", vp.cfg.CurrentQuality, vp.cfg.CurrentDub)
	}

	return nil, fmt.Errorf("озвучка '%s' не найдена", vp.cfg.CurrentDub)
}

func (vp *videoPlayer) SelectDub(dub string) error {
	if _, exists := vp.Videos[dub]; !exists {
		return fmt.Errorf("озвучка '%s' не найдена", dub)
	}
	vp.cfg.CurrentDub = dub

	qualities, err := vp.GetQualities(dub)
	if err != nil {
		return err
	}

	if vp.cfg.CurrentQuality == 0 {
		vp.cfg.CurrentQuality = qualities[0]
		return nil
	}

	vp.cfg.CurrentQuality = vp.getClosestQuality(vp.cfg.CurrentQuality, qualities)

	return nil
}

func (vp *videoPlayer) SelectQuality(quality int) error {
	if vp.cfg.CurrentDub == "" {
		return fmt.Errorf("озвучка не выбрана")
	}
	qualities, _ := vp.GetQualities(vp.cfg.CurrentDub)
	vp.cfg.CurrentQuality = vp.getClosestQuality(quality, qualities)
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
			logger.ErrorLog.Printf("не удача mpv на %d попытке: process == nil \n", i+1)
			continue
		}

		if err := cmd.Wait(); err != nil {
			logger.ErrorLog.Printf("не удача mpv на %d попытке %s\n", i+1, err)
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
