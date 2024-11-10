package promptsearch

import (
	"anicliru/internal/cli/ansi"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func PromptSearchInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(ansi.ColorPrompt + "Поиск по названию: " + ansi.ColorReset)
	titleName, err := reader.ReadString('\n')
	titleName = strings.TrimSuffix(titleName, "\n")
	return titleName, err
}
