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
	fmt.Println("Поиск аниме по вашему запросу...")
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
		isExitOnQuit, cursor := ac.watchMenu.NewPrompt(actionSlice, fmt.Sprintf("Серия %d. Выберите действие:", cursorEpisode))
		if isExitOnQuit {
			return
		}
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
	cmd := exec.Command("mpv", " --cache=on", url)
	if err := cmd.Start(); err != nil {
		panic(err)
	}
}
