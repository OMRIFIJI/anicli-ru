package video

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
)

type VideoPlayer struct {
	Links          map[string]map[int]string
	CurrentDub     string
	CurrentQuality int
}

func NewVideoPlayer(links map[string]map[int]string) *VideoPlayer {
	return &VideoPlayer{
		Links:          links,
		CurrentQuality: 0,
	}
}

func (vp *VideoPlayer) SetLinks(newLinks map[string]map[int]string) error {
	vp.Links = newLinks

	if _, exists := vp.Links[vp.CurrentDub]; !exists {
		vp.CurrentDub = ""
		return errors.New("Выбранная озвучка больше не доступна")
	}

	qualities, _ := vp.GetQualities(vp.CurrentDub)
	vp.CurrentQuality = vp.getClosestQuality(vp.CurrentQuality, qualities)

	return nil
}

func (vp *VideoPlayer) GetDubs() []string {
	dubs := make([]string, 0, len(vp.Links))
	for dub := range vp.Links {
		dubs = append(dubs, dub)
	}

	sort.Strings(dubs)
	return dubs
}

func (vp *VideoPlayer) GetQualities(dub string) ([]int, error) {
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

func (vp *VideoPlayer) GetLink() (string, error) {
	if vp.CurrentDub == "" {
		return "", errors.New("Озвучка не выбрана")
	}

	if qualities, exists := vp.Links[vp.CurrentDub]; exists {
		if link, exists := qualities[vp.CurrentQuality]; exists {
			return link, nil
		}
		return "", fmt.Errorf("Качество %d не существует для озвучки '%s'", vp.CurrentQuality, vp.CurrentDub)
	}

	return "", fmt.Errorf("Озвучка '%s' не найдена", vp.CurrentDub)
}

func (vp *VideoPlayer) SelectDub(dub string) error {
	if _, exists := vp.Links[dub]; !exists {
		return fmt.Errorf("Озвучка '%s' не найдена", dub)
	}
	vp.CurrentDub = dub

	qualities, _ := vp.GetQualities(dub)

	if vp.CurrentQuality == 0 {
		vp.CurrentQuality = qualities[0]
		return nil
	}

	vp.CurrentQuality = vp.getClosestQuality(vp.CurrentQuality, qualities)

	return nil
}

func (vp *VideoPlayer) SelectQuality(quality int) error {
	if vp.CurrentDub == "" {
		return fmt.Errorf("Озвучка не выбрана")
	}
	qualities, _ := vp.GetQualities(vp.CurrentDub)
	vp.CurrentQuality = vp.getClosestQuality(quality, qualities)
	return nil
}

func (vp *VideoPlayer) getClosestQuality(target int, qualities []int) int {
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

func (vp *VideoPlayer) StartMpv(title string) error {
	link, err := vp.GetLink()
	if err != nil {
		return err
	}

	cmd := exec.Command("mpv", "--force-media-title="+title, link)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Не удалось запустить MPV: %s.\n", err)
	}

	return nil
}

func intAbs(value int) int {
	if value < 0 {
		return (-1) * value
	}
	return value
}
