package promptselect

import (
	"context"
	"errors"
	"os"
	"sync"
	"unicode/utf8"

	"github.com/OMRIFIJI/anicli-ru/internal/cli/ansi"

	"golang.org/x/term"
)

func NewPrompt(entries []string, promptMessage string, showIndex bool) (*PromptSelect, error) {
	if len(entries) == 0 {
		return nil, errors.New("ничего не найдено")
	}
	promptCtx := promptContext{
		promptMessage: promptMessage,
		entries:       entries,
		cur:           0,
	}

	drawer, err := newDrawer(promptCtx, showIndex)
	if err != nil {
		return nil, err
	}

	ch := promptChannels{
		keyCode:  make(chan keyCode, 2),
		exitCode: make(chan exitPromptCode),
	}

	p := PromptSelect{
		promptCtx: promptCtx,
		ch:        ch,
		drawer:    drawer,
	}

	return &p, nil
}

func PrepareTerminal() (*term.State, error) {
	enterAltScreenBuf()
	ansi.HideCursor()

	fd := int(os.Stdin.Fd())

	oldTermState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	return oldTermState, err
}

func RestoreTerminal(oldTermState *term.State) {
	fd := int(os.Stdin.Fd())
	term.Restore(fd, oldTermState)

	ansi.ShowCursor()
	exitAltScreenBuf()
}

func (p *PromptSelect) SpinPrompt() (bool, int, error) {
	exitCodeValue, err := p.promptUserChoice()
	return exitCodeValue == onQuitExitCode, p.promptCtx.cur, err
}

func (p *PromptSelect) promptUserChoice() (exitPromptCode, error) {
	backgroundCtx := context.Background()
	ctx, cancel := context.WithCancelCause(backgroundCtx)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		p.spinHandleInput(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		p.drawer.spinDrawInterface(p.ch.keyCode, ctx, cancel)
	}()
	defer wg.Wait()

	// Дочитать из канала после выхода в случае ошибки
	defer func() {
		for range p.ch.keyCode {
		}
	}()

	defer cancel(nil)

	for {
		select {
		case exitCode := <-p.ch.exitCode:
			return exitCode, nil
		case <-ctx.Done():
			err := context.Cause(ctx)
			if err == context.Canceled {
				return onQuitExitCode, nil
			}
			return onErrorExitCode, err
		}
	}
}

func (p *PromptSelect) spinHandleInput(ctx context.Context) {
	defer close(p.ch.keyCode)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			keyCodeValue, err := p.readKey()
			if err != nil {
				continue
			}
			switch keyCodeValue {
			case quitKeyCode:
				p.ch.exitCode <- onQuitExitCode
				return
			case enterKeyCode:
				p.ch.exitCode <- onEnterExitCode
				return
			case upKeyCode, downKeyCode:
				p.ch.keyCode <- keyCodeValue
				p.moveCursor(keyCodeValue)
			}
		}
	}
}

func (p *PromptSelect) readKey() (keyCode, error) {
	var buf [3]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return noActionKeyCode, err
	}

	if n == 1 {
		switch buf[0] {
		case '\n', '\r':
			return enterKeyCode, nil
		case 'q', 3, 4:
			return quitKeyCode, nil
		case 'k':
			return upKeyCode, nil
		case 'j':
			return downKeyCode, nil
		case 'l':
			return enterKeyCode, nil
		}
	}
	// Обрабатывает 'й' как 'q'
	if n == 2 {
		r, _ := utf8.DecodeRune(buf[:])
		switch r {
		case 'й':
			return quitKeyCode, nil
		case 'л':
			return upKeyCode, nil
		case 'о':
			return downKeyCode, nil
		case 'д':
			return enterKeyCode, nil
		}
	}
	if n == 3 && buf[0] == 27 && buf[1] == 91 {
		switch buf[2] {
		case 65:
			return upKeyCode, nil
		case 66:
			return downKeyCode, nil
		}
	}

	return noActionKeyCode, nil
}

func (p *PromptSelect) moveCursor(keyCodeValue keyCode) {
	switch keyCodeValue {
	case downKeyCode:
		if p.promptCtx.cur < len(p.promptCtx.entries)-1 {
			p.promptCtx.cur++
		}
	case upKeyCode:
		if p.promptCtx.cur > 0 {
			p.promptCtx.cur--
		}
	}
}
