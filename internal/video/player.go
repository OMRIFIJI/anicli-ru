package video

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"syscall"
)

type videoPlayer struct {
	Links map[string]map[int]string
	cfg   videoPlayerConfig
	pid   int
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
    return vpc.CurrentQuality == 0
}

func newVideoPlayer() *videoPlayer {
	return &videoPlayer{
		cfg: videoPlayerConfig{CurrentQuality: 0},
		pid: 0,
	}
}

func (vp *videoPlayer) SetLinks(newLinks map[string]map[int]string) error {
	vp.Links = newLinks

	if _, exists := vp.Links[vp.cfg.CurrentDub]; !exists {
		vp.cfg.CurrentDub = ""
		err := &noDubError{
			Msg: "Выбранная озвучка больше не доступна. Выберите новую озвучку. ",
		}
		return err
	}

	qualities, _ := vp.GetQualities(vp.cfg.CurrentDub)
	vp.cfg.CurrentQuality = vp.getClosestQuality(vp.cfg.CurrentQuality, qualities)

	return nil
}

func (vp *videoPlayer) GetDubs() []string {
	dubs := make([]string, 0, len(vp.Links))
	for dub := range vp.Links {
		dubs = append(dubs, dub)
	}

	sort.Strings(dubs)
	return dubs
}

func (vp *videoPlayer) GetQualities(dub string) ([]int, error) {
	if qualities, exists := vp.Links[dub]; exists {
		qualityList := make([]int, 0, len(qualities))
		for quality := range qualities {
			qualityList = append(qualityList, quality)
		}

		sort.Sort(sort.Reverse(sort.IntSlice(qualityList)))
		return qualityList, nil
	}

	return nil, fmt.Errorf("озвучка '%s' не найдена", dub)
}

func (vp *videoPlayer) GetLink() (string, error) {
	if vp.cfg.CurrentDub == "" {
		return "", errors.New("Озвучка не выбрана")
	}

	if qualities, exists := vp.Links[vp.cfg.CurrentDub]; exists {
		if link, exists := qualities[vp.cfg.CurrentQuality]; exists {
			return link, nil
		}
		return "", fmt.Errorf("Качество %d не существует для озвучки '%s'", vp.cfg.CurrentQuality, vp.cfg.CurrentDub)
	}

	return "", fmt.Errorf("Озвучка '%s' не найдена", vp.cfg.CurrentDub)
}

func (vp *videoPlayer) SelectDub(dub string) error {
	if _, exists := vp.Links[dub]; !exists {
		return fmt.Errorf("Озвучка '%s' не найдена", dub)
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
		return fmt.Errorf("Озвучка не выбрана")
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

func (vp *videoPlayer) StartMpv(title string) error {
	link, err := vp.GetLink()
	if err != nil {
		return err
	}

	cmd := exec.Command("mpv", "--force-media-title="+title, link)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Не удалось запустить MPV: %s.", err)
	}

	vp.pid = cmd.Process.Pid

	return nil
}

func (vp *videoPlayer) KillMpv() error {
	if vp.pid == 0 {
		return errors.New("Попытка убить процесс MPV, не созданный этим приложением.")
	}

	syscall.Kill(vp.pid, syscall.SIGKILL)
	return nil
}

func intAbs(value int) int {
	if value < 0 {
		return (-1) * value
	}
	return value
}
