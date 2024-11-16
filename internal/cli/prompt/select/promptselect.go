package promptselect

import (
	"anicliru/internal/cli/ansi"
	"fmt"
	"os"
	"sync"
	"unicode/utf8"

	"golang.org/x/term"
)

func NewPrompt(entryNames []string, promptMessage string) (*PromptSelect, error) {
	var s PromptSelect
	if err := s.init(entryNames, promptMessage); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *PromptSelect) SpinPrompt() (bool, int, error) {
	exitCodeValue, err := s.promptUserChoice()
	return exitCodeValue == onQuitExitCode, s.promptCtx.cur, err
}

func (s *PromptSelect) init(entries []string, promptMessage string) error {
	s.promptCtx = promptContext{
		promptMessage: promptMessage,
		cur: 0,
		wg: &sync.WaitGroup{},
	}

    // Добавление нумерации
    s.promptCtx.entries = make([]string, 0, len(entries))
	for i := 0; i < len(entries); i++ {
		s.promptCtx.entries = append(s.promptCtx.entries, fmt.Sprintf("%d %s", i+1, entries[i]))
	}

	s.ch = promptChannels{
		keyCode:  make(chan keyCode),
		exitCode: make(chan exitPromptCode),
		err:      make(chan error),
	}

	drawer, err := newDrawer(s.promptCtx)
	if err != nil {
		return err
	}
	s.drawer = drawer
	return nil
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
		if s.promptCtx.cur < len(s.promptCtx.entries)-1 {
			s.promptCtx.cur++
		}
	case upKeyCode:
		if s.promptCtx.cur > 0 {
			s.promptCtx.cur--
		}
	}
}
