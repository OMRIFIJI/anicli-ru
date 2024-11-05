package clicontroller

import (
	"bufio"
    "os"
    "fmt"
    "strings"
    "anicliru/internal/cli/promptselect"
    txtclr "anicliru/internal/cli/textcolors"
)

type CLIController struct {
	TitleSelect   promptselect.PromptSelect
	EpisodeSelect promptselect.PromptSelect
}

func (c *CLIController) PromptAnimeTitleInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(txtclr.ColorPrompt + "Поиск по названию: " + txtclr.ColorReset)
	titleName, err := reader.ReadString('\n')
	titleName = strings.TrimSuffix(titleName, "\n")
	return titleName, err
}

func (c *CLIController) SearchResEmptyNotify() {
	fmt.Println(txtclr.ColorErr + "Ничего не нашлось. Попробуйте написать название точнее." + txtclr.ColorReset)
}
