package promptselect

import (
	"golang.org/x/term"
	"os"
)

type Cursor struct {
	Pos    int
	posOld int
	posMax int
}

type PromptSelect struct {
	PromptMessage string
	Cur           Cursor
	entryNames    []string
	entryCount    int
	drawer        Drawer
	entries       [][]string
	indToDraw     []int
	termSize      terminalSize
}

func (s *PromptSelect) Prompt(entryNames []string) bool {
	s.entryNames = entryNames
	s.entryCount = len(s.entryNames)
    s.setDefaultParams()

	enterAltScreenBuf()
    // Так рано defer, чтобы вернуть курсор после panic
	defer showCursor()
	defer exitAltScreenBuf()
	exitCode := s.promptUserChoice()

	return exitCode == quitCode
}

func (s *PromptSelect) setDefaultParams() {
	s.Cur = Cursor{
		Pos:    0,
		posOld: 0,
		posMax: s.entryCount - 1,
	}
	s.drawer = Drawer{
		cur: &s.Cur,
	}
	termWidth, termHeight, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}
	s.termSize = terminalSize{
		width:  termWidth,
		height: termHeight,
	}
}

func (s *PromptSelect) promptUserChoice() int {
	s.drawer.initInterface(s.entryNames, s.termSize, s.PromptMessage)
	// Первая отрисовка
	s.drawer.drawInterface()
	for {
		key, _ := s.readKey()
		keyCode := s.handleChoiceInput(key)

		switch keyCode {
		case enterCode, quitCode:
			return keyCode
		case cursorCode:
			s.drawer.drawInterface()
		}
	}
}

func (s *PromptSelect) handleChoiceInput(key string) int {
	switch key {
	case "down":
		if s.Cur.Pos < s.Cur.posMax {
			s.Cur.posOld = s.Cur.Pos
			s.Cur.Pos++
			return cursorCode
		}
	case "up":
		if s.Cur.Pos > 0 {
			s.Cur.posOld = s.Cur.Pos
			s.Cur.Pos--
			return cursorCode
		}
	case "quit":
		return quitCode
	case "enter":
		return enterCode
	}
	// Можно сказать особый случай switch
	return continueCode
}

func (s *PromptSelect) readKey() (string, error) {
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
