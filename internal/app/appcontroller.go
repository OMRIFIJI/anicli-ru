package app

import (
	"anicliru/internal/cli/promptselect"
	txtclr "anicliru/internal/cli/textcolors"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type AppController struct {
	TitleSelect   promptselect.PromptSelect
	EpisodeSelect promptselect.PromptSelect
	watchMenu     promptselect.PromptSelect
}

func (ac *AppController) PromptAnimeTitleInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(txtclr.ColorPrompt + "Поиск по названию: " + txtclr.ColorReset)
	titleName, err := reader.ReadString('\n')
	titleName = strings.TrimSuffix(titleName, "\n")
	return titleName, err
}

func (ac *AppController) SearchResEmptyNotify() {
	fmt.Println(txtclr.ColorErr + "Ничего не нашлось. Попробуйте написать название точнее." + txtclr.ColorReset)
}

func (ac *AppController) WatchMenuSpin(episodeLinks map[string]map[string]string, cursorEpisode int) {
	episodeCount := len(episodeLinks)
	go ac.startVideo(episodeLinks, cursorEpisode)
	// Рисует экран

	actionSlice := []string{"Следующая серия", "Предыдущая серия", "Выйти"}
	for {
		ac.watchMenu = promptselect.PromptSelect{
			PromptMessage: fmt.Sprintf("Серия %d. Выберите действие:", cursorEpisode),
		}
		isExitOnQuit := ac.watchMenu.Prompt(actionSlice)
		if isExitOnQuit {
			return
		}
		cursor := ac.watchMenu.Cur.Pos
		switch cursor {
		case 0:
			cursorEpisode++
			if cursorEpisode > episodeCount {
				return
			}
		case 1:
			if cursorEpisode > 1 {
				cursorEpisode--
			}
		case 2:
			return
		}

		go ac.startVideo(episodeLinks, cursorEpisode)
	}
}

func (ac *AppController) startVideo(episodeLinks map[string]map[string]string, cursorEpisode int) {
	url := episodeLinks[strconv.Itoa(cursorEpisode)]["FHD"]
	cmd := exec.Command("mpv", url)
	if err := cmd.Start(); err != nil {
		panic(err)
	}
}
