package promptselect

import (
    "anicliru/internal/cli/ansi"
	"golang.org/x/term"
	"os"
	"sync"
	"unicode/utf8"
)

func (s *PromptSelect) NewPrompt(entryNames []string, promptMessage string) (bool, int, error) {
	s.init(entryNames, promptMessage)

	exitCodeValue, err := s.promptUserChoice()

	return exitCodeValue == onQuitExitCode, s.promptCtx.cur.pos, err
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

	s.ch = promptChannels{
		keyCode:  make(chan keyCode),
		exitCode: make(chan exitPromptCode),
		err:      make(chan error),
	}

	s.drawer = Drawer{}
	s.drawer.newDrawer(s.promptCtx)
}

func (s *PromptSelect) promptUserChoice() (exitPromptCode, error) {
	ansi.EnterAltScreenBuf()
	defer ansi.ExitAltScreenBuf()

	oldTermState, err := term.MakeRaw(0)
	if err != nil {
        return onErrorExitCode, err
	}
	defer term.Restore(0, oldTermState)

	ansi.HideCursor()
	defer ansi.ShowCursor()

	s.promptCtx.wg.Add(1)

	go s.spinHandleInput()

	go s.drawer.spinDrawInterface(s.ch.keyCode, s.ch.err)

	select {
	case err := <-s.ch.err:
        return onErrorExitCode, err
	case exitCode := <-s.ch.exitCode:
		return exitCode, err
	}
}

func (s *PromptSelect) spinHandleInput() {
	for {
		keyCodeValue, err := s.readKey()
		if err != nil {
			s.ch.err <- err
		}
		s.ch.keyCode <- keyCodeValue

		switch keyCodeValue {
		case quitKeyCode:
			s.promptCtx.wg.Wait()
			s.ch.exitCode <- onQuitExitCode
			return
		case enterKeyCode:
			s.promptCtx.wg.Wait()
			s.ch.exitCode <- onEnterExitCode
			return
		case upKeyCode, downKeyCode:
			s.moveCursor(keyCodeValue)
		}
	}
}

func (s *PromptSelect) readKey() (keyCode, error) {
	// Терминал в raw mode
	var buf [3]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return noActionKeyCode, err
	}

	if n == 1 && (buf[0] == '\n' || buf[0] == '\r') {
		return enterKeyCode, nil
	}
	if n == 1 && buf[0] == 'q' {
		return quitKeyCode, nil
	}
	// Обрабатывает 'й' как 'q'
	if n == 2 {
		r, _ := utf8.DecodeRune(buf[:])
		if r == 'й' {
			return quitKeyCode, nil
		}
	}
	if n >= 3 && buf[0] == 27 && buf[1] == 91 {
		switch buf[2] {
		case 65:
			return upKeyCode, nil
		case 66:
			return downKeyCode, nil
		}
	}

	return noActionKeyCode, nil
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
