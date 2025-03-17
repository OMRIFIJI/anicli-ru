package video

import (
	promptselect "github.com/OMRIFIJI/anicli-ru/internal/cli/prompt/select"
	"strconv"
)

const (
	nextEpisode     string = "Следующая серия"
	previousEpisode string = "Предыдущая серия"
	replay          string = "Перезапустить плеер"
	changeDub       string = "Выбрать другую озвучку"
	changeQuality   string = "Изменить качество видео"
	exitPlayer      string = "Выход"
)

type videoSelector struct {
	menuOptions []string
}

func newSelector() *videoSelector {
	v := &videoSelector{
		menuOptions: []string{
			nextEpisode,
			previousEpisode,
			replay,
			changeDub,
			changeQuality,
			exitPlayer,
		},
	}
	return v
}

func (vs *videoSelector) selectDub(promptMessage string, player *videoPlayer) (bool, error) {
	dubEntries := player.GetDubs()
	prompt, err := promptselect.NewPrompt(dubEntries, promptMessage, false)
	if err != nil {
		return false, err
	}

	isExitOnQuit, cur, err := prompt.SpinPrompt()
	if err != nil {
		return false, err
	}

	err = player.SelectDub(dubEntries[cur])
	if err != nil {
		return false, err
	}

	return isExitOnQuit, err
}

func (vs *videoSelector) selectQuality(promptMessage string, player *videoPlayer) (bool, error) {
	qualities, err := player.GetQualities(player.cfg.Dub)
	if err != nil {
		return false, err
	}

	var qualityEntries []string
	for _, quality := range qualities {
		qualityEntries = append(qualityEntries, strconv.Itoa(quality))
	}

	prompt, err := promptselect.NewPrompt(qualityEntries, promptMessage, false)
	if err != nil {
		return false, err
	}

	isExitOnQuit, cur, err := prompt.SpinPrompt()
	if err != nil {
		return false, err
	}

	err = player.SelectQuality(qualities[cur])
	if err != nil {
		return false, err
	}

	return isExitOnQuit, err
}

func (vs *videoSelector) selectMenuOption(promptMessage string) (string, bool, error) {
	prompt, err := promptselect.NewPrompt(vs.menuOptions, promptMessage, false)
	if err != nil {
		return "", false, err
	}

	isExitOnQuit, cur, err := prompt.SpinPrompt()
	if err != nil {
		return "", false, err
	}

	return vs.menuOptions[cur], isExitOnQuit, nil
}
