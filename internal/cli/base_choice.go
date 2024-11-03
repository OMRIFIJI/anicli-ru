package cli

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

type BaseChoiceHandler struct {
	cursor          int
	foundAnimeInfo  []string
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

func (c *BaseChoiceHandler) PromptSearchRes(foundAnimeInfo []string) bool {
	c.foundAnimeInfo = foundAnimeInfo
	c.foundAnimeCount = len(c.foundAnimeInfo)
	c.enterAltScreenBuf()

	isUserChoiceQuit := c.promptUserChoice()

	c.exitAltScreenBuf()

	return isUserChoiceQuit
}

func (c *BaseChoiceHandler) promptUserChoice() bool {
	for {
		c.printOptions()
		key, _ := readKey()
		isUserChoiceQuit := c.handleChoiceInput(key)
		if isUserChoiceQuit {
			return true
		}
	}
}

func (c *BaseChoiceHandler) handleChoiceInput(key string) bool {
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

func (c *BaseChoiceHandler) enterAltScreenBuf() {
	fmt.Print("\033[?1049h")
}

func (c *BaseChoiceHandler) clearScreen() {
	fmt.Print("\033[H\033[J")
}

func (c *BaseChoiceHandler) exitAltScreenBuf() {
	fmt.Print("\033[?1049l")
}

func (c *BaseChoiceHandler) printOptions() {
	c.clearScreen()
	fmt.Printf("%s%s%s\n", yellow, "Выберите аниме из списка:", colorReset)
	for i, option := range c.foundAnimeInfo {
		if i == c.cursor {
			selectedEntry := c.selectedEntryString(i, option)
			fmt.Println(selectedEntry)
		} else {
			fmt.Printf("%s %s  %d %s\n", highlightBg, highlightReset, i+1, option)
		}
	}
	fmt.Println("\nНажмите 'Enter' чтобы выбрать, или 'q' выхода.")
}

func (c *BaseChoiceHandler) selectedEntryString(i int, option string) string {
	var b strings.Builder
	fmt.Fprintf(
		&b,
		"%s%s  %d %s",
		highlightBg,
		highlightFg,
		i+1,
		option,
	)
	// Дополнительные пробелы, чтобы серый фон был до конца строки
	termWidth, _, _ := term.GetSize(0)
	additionalSpaces := termWidth - (utf8.RuneCountInString(option) + len(strconv.Itoa(i)) + 3)
	for additionalSpaces < 0 {
		additionalSpaces += termWidth
	}
	b.WriteString(strings.Repeat(" ", additionalSpaces))
	b.WriteString(highlightReset)
	return b.String()
}
