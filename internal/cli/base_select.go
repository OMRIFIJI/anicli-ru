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
		entry = c.fmtEntryString(entry, termWidth, i == c.cursor)
		fmt.Print(entry)
	}

	fmt.Printf("└%s┘\n", strings.Repeat("─", termWidth-2))
}

func (c *BaseSelectHandler) fmtEntryString(entry string, termWidth int, active bool) string {
	entryRune := []rune(entry)
	entryRuneLen := len(entryRune)
	var b strings.Builder
	// Сколько чистого текста вмещается в табличку
	altScreenWidth := termWidth - 7

	// Записываем весь entry в одну строку, если можем
	if entryRuneLen <= altScreenWidth {
		extraSpaces := altScreenWidth - entryRuneLen
		c.formatLine(
			&b,
			string(entryRune[:entryRuneLen]),
			fmtOpts{
				active:      active,
				extraSpaces: extraSpaces,
				LeftPadding: 2,
			},
		)
		return b.String()
	} else {
        c.formatLine(
			&b,
			string(entryRune[:altScreenWidth]),
			fmtOpts{
				active:      active,
				extraSpaces: 0,
				LeftPadding: 2,
			},
		)
	}

	// Остальные строки entry кроме последней
	left := altScreenWidth
	right := left + altScreenWidth - 2
	for right < entryRuneLen {
        c.formatLine(
			&b,
			string(entryRune[left:right]),
			fmtOpts{
				active:      active,
				extraSpaces: 0,
				LeftPadding: 4,
			},
		)
		left += altScreenWidth - 2
		right += altScreenWidth - 2
	}

	// Последняя строка, надо снова заполнить пробелами
	extraSpaces := altScreenWidth - 2 - (entryRuneLen - left)
    c.formatLine(
        &b,
        string(entryRune[left:]),
        fmtOpts{
            active:      active,
            extraSpaces: extraSpaces,
            LeftPadding: 4,
        },
    )

	return b.String()
}

type fmtOpts struct {
	active      bool
	extraSpaces int
	LeftPadding int
}

func (c *BaseSelectHandler) formatLine(b *strings.Builder, entryLine string, opts fmtOpts) {
	b.WriteString("│ ")

	if opts.active {
		fmt.Fprintf(b, "%s%s▌%s", highlightBg, highlightCursor, highlightFg)
		fmt.Fprintf(b, "%s%s", strings.Repeat(" ", opts.LeftPadding), highlightFg)
	} else {
		fmt.Fprintf(b, "%s %s", highlightBg, highlightBgReset)
		b.WriteString(strings.Repeat(" ", opts.LeftPadding))
	}

	b.WriteString(entryLine)
	if opts.extraSpaces > 0 {
		b.WriteString(strings.Repeat(" ", opts.extraSpaces))
	}
	if opts.active {
		b.WriteString(highlightBgReset)
	}
	b.WriteString(" │\n")
}
