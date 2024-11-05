package cli

import (
	"errors"
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
)

type Cursor struct {
	Pos    int
	posOld int
	state  int
	posMax int
}

// func (c *Cursor) updateState() {
// }

type terminalSize struct {
	width  int
	height int
}

type SelectPrompt struct {
	Cur        Cursor
	entryNames []string
	entryCount int
	entries    [][]string
	indToDraw  []int
	term       terminalSize
}

func (s *SelectPrompt) PromptSearchRes(entryNames []string) int {
	s.entryNames = entryNames
	s.entryCount = len(s.entryNames)

	// Стандартные значения
	s.Cur = Cursor{
		Pos:    0,
		posOld: 0,
		state:  cursorStateNormal,
		posMax: s.entryCount - 1,
	}

	s.enterAltScreenBuf()
	exitCode := s.promptUserChoice()
	defer s.exitAltScreenBuf()
	defer s.showCursor()

	return exitCode
}

func (s *SelectPrompt) promptUserChoice() int {
	s.initInterface()
	for {
		s.drawInterface()
		key, _ := readKey()
		exitCode := s.handleChoiceInput(key)
		if (exitCode == EnterCode) || (exitCode == QuitCode) {
			return exitCode
		}
	}
}

func (s *SelectPrompt) handleChoiceInput(key string) int {
	switch key {
	case "down":
		if s.Cur.Pos < s.Cur.posMax {
			s.Cur.posOld = s.Cur.Pos
			s.Cur.Pos++
		}
	case "up":
		if s.Cur.Pos > 0 {
			s.Cur.posOld = s.Cur.Pos
			s.Cur.Pos--
		}
	case "quit":
		return QuitCode
	case "enter":
		return EnterCode
	}
	// Можно сказать особый случай switch
	return continueCode
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

	if n == 1 && (buf[0] == '\n' || buf[0] == '\r') {
		return "enter", nil
	}
	if n == 1 && buf[0] == 'q' {
		return "quit", nil
	}
	// Обрабатывает 'й' как 'q'
	if n == 2 && buf[0] == 0xd0 && buf[1] == 0x99 {
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

func (s *SelectPrompt) enterAltScreenBuf() {
	fmt.Print("\033[?1049h")
}

func (s *SelectPrompt) exitAltScreenBuf() {
	fmt.Print("\033[?1049l")
}

func (s *SelectPrompt) clearScreen() {
	fmt.Print("\033[H\033[J")
}

func (s *SelectPrompt) hideCursor() {
	fmt.Print("\033[?25l")
}

func (s *SelectPrompt) showCursor() {
	fmt.Print("\033[?25h")
}

func (s *SelectPrompt) initInterface() {
	termWidth, termHeight, err := term.GetSize(0)

	if err != nil {
		panic(err)
	}

	if termWidth < 10 {
		panic(errors.New("Your screen is too small! Less than 10 characters."))
	}

	s.term = terminalSize{
		width:  termWidth,
		height: termHeight,
	}

	for i, entry := range s.entryNames {
		//entryDecorated := s.decorateEdge(entry, termWidth)
		s.entries = append(s.entries, s.fitEntryLines(entry, s.term.width, i == s.Cur.Pos))
	}

	s.indToDraw = make([]int, 0, s.term.height-3)

	linesCount := 0
	for i, entry := range s.entries {
		s.indToDraw = append(s.indToDraw, i)
		linesCount += len(entry)
		if linesCount > s.term.height-3 {
			break
		}

	}

}
func (s *SelectPrompt) drawInterface() {
	s.clearScreen()
	s.hideCursor()

	fmt.Printf("%s%s%s\n", colorPrompt, "Выберите аниме из списка:", colorReset)
	fmt.Printf("┌%s┐\n", strings.Repeat("─", s.term.width-2))

	s.drawEntries()

	fmt.Printf("└%s┘\n", strings.Repeat("─", s.term.width-2))
}

func (s *SelectPrompt) drawEntries() {
	linesCount := 0
	for _, i := range s.indToDraw {
		for _, entry := range s.entries[i] {
			fmt.Print(entry)
			linesCount++
			if linesCount == s.term.height-3 {
				return
			}
		}
	}

}

func (s *SelectPrompt) fitEntryLines(entry string, termWidth int, active bool) []string {
	var entryStrings []string
	entryRune := []rune(entry)
	entryRuneLen := len(entryRune)
	// Сколько чистого текста вмещается в табличку
	altScreenWidth := termWidth - 7

	// Записываем весь entry в одну строку, если можем
	if entryRuneLen <= altScreenWidth {
		extraSpaces := altScreenWidth - entryRuneLen
		entryStrings = append(
			entryStrings,
			s.formatLine(
				string(entryRune[:entryRuneLen]),
				fmtOpts{
					active:      active,
					extraSpaces: extraSpaces,
					LeftPadding: 2,
				},
			),
		)
		return entryStrings
	} else {
		entryStrings = append(
			entryStrings,
			s.formatLine(
				string(entryRune[:altScreenWidth]),
				fmtOpts{
					active:      active,
					extraSpaces: 0,
					LeftPadding: 2,
				},
			),
		)
	}

	// Остальные строки entry кроме последней
	left := altScreenWidth
	right := left + altScreenWidth - 2
	newLinesCount := 1
	for right < entryRuneLen {
		entryStrings = append(entryStrings,
			s.formatLine(
				string(entryRune[left:right]),
				fmtOpts{
					active:      active,
					extraSpaces: 0,
					LeftPadding: 4,
				},
			),
		)
		newLinesCount += 1
		left += altScreenWidth - 2
		right += altScreenWidth - 2
	}

	// Последняя строка, надо снова заполнить пробелами
	extraSpaces := altScreenWidth - 2 - (entryRuneLen - left)
	entryStrings = append(entryStrings,
		s.formatLine(
			string(entryRune[left:]),
			fmtOpts{
				active:      active,
				extraSpaces: extraSpaces,
				LeftPadding: 4,
			},
		),
	)

	return entryStrings
}

type fmtOpts struct {
	active      bool
	extraSpaces int
	LeftPadding int
}

func (s *SelectPrompt) formatLine(entryLine string, opts fmtOpts) string {
	var b strings.Builder
	b.WriteString("│ ")

	if opts.active {
		fmt.Fprintf(&b, "%s%s▌%s", highlightBg, highlightCursor, highlightFg)
		fmt.Fprintf(&b, "%s%s", strings.Repeat(" ", opts.LeftPadding), highlightFg)
	} else {
		fmt.Fprintf(&b, "%s %s", highlightBg, highlightBgReset)
		b.WriteString(strings.Repeat(" ", opts.LeftPadding))
	}

	b.WriteString(strings.TrimSpace(entryLine))
	if opts.extraSpaces > 0 {
		b.WriteString(strings.Repeat(" ", opts.extraSpaces))
	}
	if opts.active {
		b.WriteString(highlightBgReset)
	}
	b.WriteString(" │\n")
	return b.String()
}
