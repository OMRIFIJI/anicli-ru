package promptselect

import (
	"golang.org/x/term"
	"os"
	"sync"
	"unicode/utf8"
)

func (s *PromptSelect) NewPrompt(entryNames []string, promptMessage string) (bool, int) {
	s.init(entryNames, promptMessage)

	exitCodeValue := s.promptUserChoice()

	return exitCodeValue == onQuitExitCode, s.promptCtx.cur.pos
}

func (s *PromptSelect) init(entryNames []string, promptMessage string) {
	s.promptCtx = promptContext{
		promptMessage: promptMessage,
		entries:       entryNames,
		cur: &Cursor{
			pos:    0,
			posMax: len(entryNames) - 1,
		},
		wg: &sync.WaitGroup{},
	}

	s.drawer = Drawer{}
	s.drawer.newDrawer(s.promptCtx)
}

func (s *PromptSelect) promptUserChoice() exitPromptCode {
	enterAltScreenBuf()
	defer exitAltScreenBuf()

	oldTermState, err := term.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer term.Restore(0, oldTermState)

	hideCursor()
	defer showCursor()

	s.promptCtx.wg.Add(1)
	keyCodeChan := make(chan keyCode)
	exitCodeChan := make(chan exitPromptCode)
	go s.spinHandleInput(keyCodeChan, exitCodeChan)

	errChan := make(chan error, 2)
	go s.drawer.spinDrawInterface(keyCodeChan, errChan)

	select {
	case err := <-errChan:
		panic(err)
	case exitCode := <-exitCodeChan:
		return exitCode
	}
}

func (s *PromptSelect) spinHandleInput(keyCodeChan chan keyCode, exitCodeChan chan exitPromptCode) {
	for {
		keyCodeValue := s.readKey()
		keyCodeChan <- keyCodeValue

		switch keyCodeValue {
		case quitKeyCode:
			s.promptCtx.wg.Wait()
			exitCodeChan <- onQuitExitCode
			return
		case enterKeyCode:
			s.promptCtx.wg.Wait()
			exitCodeChan <- onEnterExitCode
			return
		case upKeyCode, downKeyCode:
			s.moveCursor(keyCodeValue)
		}
	}
}

func (s *PromptSelect) readKey() keyCode {
	// Терминал в raw mode
	var buf [3]byte
	n, _ := os.Stdin.Read(buf[:])

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

	return noActionKeyCode
}

func (s *PromptSelect) moveCursor(keyCodeValue keyCode) {
	switch keyCodeValue {
	case downKeyCode:
		if s.promptCtx.cur.pos < s.promptCtx.cur.posMax {
			s.promptCtx.cur.pos++
		}
	case upKeyCode:
		if s.promptCtx.cur.pos > 0 {
			s.promptCtx.cur.pos--
		}
	}
}
