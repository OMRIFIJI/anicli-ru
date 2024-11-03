package cli

import (
	"bufio"
    "os"
    "fmt"
    "strings"
)

type CLI struct {
	titleChoice   BaseChoiceHandler
	episodeChoice BaseChoiceHandler
}

func (c *CLI) PromptAnimeTitle() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(yellow + "Поиск по названию: " + colorReset)
	titleName, err := reader.ReadString('\n')
	titleName = strings.TrimSuffix(titleName, "\n")
	return titleName, err
}

func (c *CLI) SearchResEmptyNotify() {
	fmt.Println(red + "Ничего не нашлось. Попробуйте написать название точнее." + colorReset)
}
