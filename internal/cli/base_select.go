package cli

import (
	"errors"
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
)

type BaseSelectHandler struct {
	cursor          int
	foundAnimeInfo  []string
	foundAnimeCount int
}

func (c *BaseSelectHandler) PromptSearchRes(foundAnimeInfo []string) bool {
	c.foundAnimeInfo = foundAnimeInfo
	c.foundAnimeCount = len(c.foundAnimeInfo)
	c.enterAltScreenBuf()

	isUserSelectQuit := c.promptUserChoice()
	defer c.showCursor()
	c.exitAltScreenBuf()

	return isUserSelectQuit
}

func (c *BaseSelectHandler) promptUserChoice() bool {
	for {
		c.drawInterface()
		key, _ := readKey()
		isUserSelectQuit := c.handleChoiceInput(key)
		if isUserSelectQuit {
			return true
		}
	}
}

func (c *BaseSelectHandler) handleChoiceInput(key string) bool {
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

func (c *BaseSelectHandler) enterAltScreenBuf() {
	fmt.Print("\033[?1049h")
}

func (c *BaseSelectHandler) exitAltScreenBuf() {
	fmt.Print("\033[?1049l")
}

func (c *BaseSelectHandler) clearScreen() {
	fmt.Print("\033[H\033[J")
}

func (c *BaseSelectHandler) hideCursor() {
	fmt.Print("\033[?25l")
}

func (c *BaseSelectHandler) showCursor() {
	fmt.Print("\033[?25h")
}

func (c *BaseSelectHandler) drawInterface() {
	c.clearScreen()
	c.hideCursor()

	termWidth, _, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}
	if termWidth < 10 {
		panic(errors.New("Your screen is too small! Less than 10 characters."))
	}

	fmt.Printf("%s%s%s\n", colorPrompt, "Выберите аниме из списка:", colorReset)
	fmt.Printf("┌%s┐\n", strings.Repeat("─", termWidth-2))

	for i, entry := range c.foundAnimeInfo {
		//entryDecorated := c.decorateEdge(entry, termWidth)
		if i == c.cursor {
			entry = c.activeEntryString(entry, termWidth)
		} else {
			entry = fmt.Sprintf("│ %s %s  %s\n", highlightBg, highlightBgReset, entry)
		}
		fmt.Print(entry)
	}
}

func (c *BaseSelectHandler) activeEntryString(entry string, termWidth int) string {
	entryRune := []rune(entry)
	entryRuneLen := len(entryRune)
	var b strings.Builder
	// Тут идёт хитрый трюк
	// Обрабатываем, когда entry не помещается в одну строку на экране
	// У первой строки отступ меньше, поэтому она отдельно от цикла пойдёт
	fmt.Fprintf(&b, "│ %s%s▌  %s", highlightBg, highlightCursor, highlightFg)
	altScreenWidth := termWidth - 7

	// Записываем весь entry в одну строку, если можем
	if entryRuneLen <= altScreenWidth {
		b.WriteString(string(entryRune[:entryRuneLen]))

	    extraSpaces := altScreenWidth - entryRuneLen
        b.WriteString(strings.Repeat(" ", extraSpaces))
        
        fmt.Fprintf(&b, "%s │\n", highlightBgReset)

        return b.String()
	} else {
		b.WriteString(string(entryRune[:altScreenWidth]))
        fmt.Fprintf(&b, "%s │\n", highlightBgReset)
	}

	// Остальные строки entry кроме последней
	left := altScreenWidth
	right := left + altScreenWidth - 2
	for right < entryRuneLen {
		fmt.Fprintf(&b, "│ %s%s", strings.Repeat(" ", 5), highlightBg)
		fmt.Fprintf(&b, "%s%s │\n", string(entryRune[left:right]), highlightBgReset)
		left += altScreenWidth - 2
		right += altScreenWidth - 2
	}

	fmt.Fprintf(&b, "│ %s%s", strings.Repeat(" ", 5), highlightBg)

	b.WriteString(string(entryRune[left:]))
	// Последняя строка, надо снова заполнить пробелами
    extraSpaces := altScreenWidth - 2 - (entryRuneLen - left)
	b.WriteString(strings.Repeat(" ", extraSpaces))
	b.WriteString(highlightBgReset)
	b.WriteString(" │\n")
	return b.String()
}
