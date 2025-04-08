package promptsearch

import (
	"bufio"
	"fmt"
	"github.com/OMRIFIJI/anicli-ru/internal/cli/ansi"
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
