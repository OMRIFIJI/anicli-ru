package promptselect

import (
	"anicliru/internal/cli/ansi"
	"context"
	"os"
	"sync"
	"unicode/utf8"

	"golang.org/x/term"
)

func NewPrompt(entryNames []string, promptMessage string, showIndex bool) (*PromptSelect, error) {
	var p PromptSelect
	if err := p.newBase(entryNames, promptMessage, showIndex); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *PromptSelect) SpinPrompt() (bool, int, error) {
	exitCodeValue, err := p.promptUserChoice()
	return exitCodeValue == onQuitExitCode, p.promptCtx.cur, err
}

func (p *PromptSelect) newBase(entries []string, promptMessage string, showIndex bool) error {
	promptCtx := promptContext{
		promptMessage: promptMessage,
		entries:       entries,
		cur:           0,
	}

	p.promptCtx = promptCtx

	p.ch = promptChannels{
		keyCode:  make(chan keyCode),
		exitCode: make(chan exitPromptCode),
	}

	drawer, err := newDrawer(promptCtx, showIndex)
	if err != nil {
		return err
	}
	p.drawer = drawer

	return nil
}

func (p *PromptSelect) promptUserChoice() (exitPromptCode, error) {
	ansi.EnterAltScreenBuf()
	defer ansi.ExitAltScreenBuf()

	oldTermState, err := term.MakeRaw(0)
	if err != nil {
		return onErrorExitCode, err
	}
	defer term.Restore(0, oldTermState)

	ansi.HideCursor()
	defer ansi.ShowCursor()

	backgroundCtx := context.Background()
	ctx, cancel := context.WithCancelCause(backgroundCtx)

    wg := sync.WaitGroup{}

    wg.Add(1)
	go func() {
        defer wg.Done()
        p.spinHandleInput(ctx, cancel)
    }()

    wg.Add(1)
	go func() {
        defer wg.Done()
        p.drawer.spinDrawInterface(p.ch.keyCode, ctx, cancel)
    }()
    defer wg.Wait()
    defer cancel(nil)

	for {
		select {
		case exitCode := <-p.ch.exitCode:
			return exitCode, nil
		case <-ctx.Done():
			return onErrorExitCode, context.Cause(ctx)
		}
	}
}

func (p *PromptSelect) spinHandleInput(ctx context.Context, cancel context.CancelCauseFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			keyCodeValue, err := p.readKey()
			if err != nil {
				cancel(err)
                return
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
