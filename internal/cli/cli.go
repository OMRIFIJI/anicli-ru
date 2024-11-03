package cli

import (
	"anicliru/internal/api"
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

type CLI struct {
	cursor          int
	foundAnimeInfo  []api.AnimeInfo
	foundAnimeCount int
}

const (
	colorReset     = "\033[0m"
	yellow         = "\033[33m"
	red            = "\033[31m"
	highlightBg    = "\033[48;5;235m"
	highlightFg    = "\033[37m"
	highlightReset = "\033[0m"
)

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

func (c *CLI) PromptSearchRes(foundAnimeInfo []api.AnimeInfo) bool {
	c.foundAnimeInfo = foundAnimeInfo
	c.foundAnimeCount = len(c.foundAnimeInfo)
	c.enterAltScreenBuf()

	isUserChoiceQuit := c.promptUserChoice()

	c.exitAltScreenBuf()

	return isUserChoiceQuit
}

func (c *CLI) promptUserChoice() bool {
	for {
		c.printOptions()
		key, _ := readKey()
		isUserChoiceQuit := c.handleChoiceInput(key)
		if isUserChoiceQuit {
			return true
		}
	}
}

func (c *CLI) handleChoiceInput(key string) bool {
	switch key {
	case "down":
		if c.cursor < c.foundAnimeCount-1 {
			c.cursor++
		}
	case "up":
		if c.cursor > 0 {
			c.cursor--
		}

	}
	// Можно сказать особый случай switch
	return key == "quit"
}

func readKey() (string, error) {
	// Терминал в raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	var buf [3]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return "", err
	}

	if n == 1 && buf[0] == '\n' {
		return "enter", nil
	}
	if n == 1 && buf[0] == 'q' {
		return "quit", nil
	}

	if n >= 3 && buf[0] == 27 && buf[1] == 91 {
		switch buf[2] {
		case 65:
			return "up", nil
		case 66:
			return "down", nil
		}
	}
	return "", nil
}

func (c *CLI) enterAltScreenBuf() {
	fmt.Print("\033[?1049h")
}

func (c *CLI) clearScreen() {
	fmt.Print("\033[H\033[J")
}

func (c *CLI) exitAltScreenBuf() {
	fmt.Print("\033[?1049l")
}

func (c *CLI) printOptions() {
	c.clearScreen()
	fmt.Printf("%s%s%s\n", yellow, "Выберите аниме из списка:", colorReset)
	for i, option := range c.foundAnimeInfo {
		if i == c.cursor {
			selectedEntry := c.selectedEntryString(i, option)
			fmt.Println(selectedEntry)
		} else {
			fmt.Printf("  %d %s\n", i+1, option.Names.RU)
		}
	}
	fmt.Println("\nНажмите 'Enter' чтобы выбрать, или 'q' выхода.")
}

func (c *CLI) selectedEntryString(i int, option api.AnimeInfo) string {
	var b strings.Builder
	fmt.Fprintf(
		&b,
		"%s%s %d %s",
		highlightBg,
		highlightFg,
		i+1,
		option.Names.RU,
	)
    // Дополнительные пробелы, чтобы серый фон был до конца строки
	termWidth, _, _ := term.GetSize(0)
	additionalSpaces := termWidth - (utf8.RuneCountInString(option.Names.RU) + len(strconv.Itoa(i)) + 2)
	for additionalSpaces < 0 {
		additionalSpaces += termWidth
	}
	b.WriteString(strings.Repeat(" ", additionalSpaces))
    b.WriteString(highlightReset)
	return b.String()
}
