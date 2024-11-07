package promptselect

import (
	"golang.org/x/term"
	"os"
	"os/signal"
	"syscall"
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

func (s *PromptSelect) Prompt(entryNames []string) bool {
	s.entryNames = entryNames
	s.setDefaultParams()

	enterAltScreenBuf()
	defer exitAltScreenBuf()

	oldState, err := term.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer term.Restore(0, oldState)

	defer showCursor()
	exitCodeValue := s.promptUserChoice()

	return exitCodeValue == onQuitExitCode
}

func (s *PromptSelect) setDefaultParams() {
	s.Cur = Cursor{
		Pos:    0,
		posOld: 0,
		posMax: len(s.entryNames) - 1,
	}
	s.drawer = Drawer{
		cur: &s.Cur,
	}
	termWidth, termHeight, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	s.termSize = terminalSize{
		width:  termWidth,
		height: termHeight,
	}
}

func (s *PromptSelect) promptUserChoice() exitPromptCode {
	// Первая отрисовка
	s.drawer.initInterface(s.entryNames, s.termSize, s.PromptMessage)
	s.drawer.drawInterface()

	quit := make(chan bool, 1)
	go s.redrawOnTerminalResize(quit)

	for {
		keyCodeValue := s.readKey()
		switch keyCodeValue {
		case quitKeyCode:
            quit <- true
			return onQuitExitCode
		case enterKeyCode:
		    quit <- true
			return onEnterExitCode
		case upKeyCode, downKeyCode:
			s.moveCursor(keyCodeValue)
		    s.drawer.drawInterface()
		}
	}
}

func (s *PromptSelect) redrawOnTerminalResize(quit chan bool) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGWINCH)

	for {
		<-signalChannel
		select {
		case <-quit:
			return
		default:
			s.setDefaultParams()
			s.drawer.initInterface(s.entryNames, s.termSize, s.PromptMessage)
			s.drawer.drawInterface()
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
