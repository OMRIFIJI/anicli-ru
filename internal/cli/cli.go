package cli

import (
	"bufio"
    "os"
    "fmt"
    "strings"
)

type CLIHandler struct {
	TitleSelect   SelectPrompt
	EpisodeSelect SelectPrompt
}

func (c *CLIHandler) PromptAnimeTitleInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(colorPrompt + "Поиск по названию: " + colorReset)
	titleName, err := reader.ReadString('\n')
	titleName = strings.TrimSuffix(titleName, "\n")
	return titleName, err
}

func (c *CLIHandler) SearchResEmptyNotify() {
	fmt.Println(colorErr + "Ничего не нашлось. Попробуйте написать название точнее." + colorReset)
}
