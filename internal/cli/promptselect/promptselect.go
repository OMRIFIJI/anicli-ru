package promptselect

import (
	"golang.org/x/term"
	"os"
	"unicode/utf8"
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
	entries       [][]string
	drawer        Drawer
	indToDraw     []int
	termSize      terminalSize
}

func (s *PromptSelect) NewPrompt(entryNames []string) bool {
	s.entryNames = entryNames
	s.setDefaultParams()


	exitCodeValue := s.promptUserChoice()

	return exitCodeValue == onQuitExitCode
}

func (s *PromptSelect) setDefaultParams() {
	s.Cur = Cursor{
		Pos:    0,
		posOld: 0,
		posMax: len(s.entryNames) - 1,
	}

	termWidth, termHeight, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}
	s.termSize = terminalSize{
		width:  termWidth,
		height: termHeight,
	}

	s.drawer = Drawer{}
	s.drawer.newDrawer(s.entryNames, s.termSize, s.PromptMessage, &s.Cur)
}

func (s *PromptSelect) promptUserChoice() exitPromptCode {
	keyCodeChan := make(chan keyCode, 1)
	go s.drawer.spinDrawInterface(keyCodeChan)

	for {
		keyCodeValue := s.readKey()
        keyCodeChan <- keyCodeValue

		switch keyCodeValue {
		case quitKeyCode:
			return onQuitExitCode
		case enterKeyCode:
			return onEnterExitCode
		case upKeyCode, downKeyCode:
			s.moveCursor(keyCodeValue)
		}
	}
}


func (s *PromptSelect) readKey() keyCode {
	// Терминал в raw mode
	var buf [3]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		panic(err)
	}

	if n == 1 && (buf[0] == '\n' || buf[0] == '\r') {
		return enterKeyCode
	}
	if n == 1 && buf[0] == 'q' {
		return quitKeyCode
	}
	// Обрабатывает 'й' как 'q'
	if n == 2 {
		r, _ := utf8.DecodeRune(buf[:])
		if r == 'й' {
			return quitKeyCode
		}
	}
	if n >= 3 && buf[0] == 27 && buf[1] == 91 {
		switch buf[2] {
		case 65:
			return upKeyCode
		case 66:
			return downKeyCode
		}
	}

	return continueKeyCode
}

func (s *PromptSelect) moveCursor(keyCodeValue keyCode) {
	switch keyCodeValue {
	case downKeyCode:
		if s.Cur.Pos < s.Cur.posMax {
			s.Cur.posOld = s.Cur.Pos
			s.Cur.Pos++
		}
	case upKeyCode:
		if s.Cur.Pos > 0 {
			s.Cur.posOld = s.Cur.Pos
			s.Cur.Pos--
		}
	}
}
