package promptselect

import (
	"anicliru/internal/cli/ansi"
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func NewPrompt(entries []string, promptMessage string, showIndex bool) (*PromptSelect, error) {
	if len(entries) == 0 {
		return nil, errors.New("Ничего не найдено")
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
	ansi.EnterAltScreenBuf()
	ansi.HideCursor()

	fd := int(os.Stdin.Fd())

	oldTermState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	return oldTermState, err
}

func RestoreTerminal(oldTermState *term.State) {
	term.Restore(0, oldTermState)
	ansi.ShowCursor()
	ansi.ExitAltScreenBuf()
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
		p.spinHandleInput(ctx, cancel)
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

func (p *PromptSelect) spinHandleInput(ctx context.Context, cancel context.CancelCauseFunc) {
	defer close(p.ch.keyCode)

	fd := int(os.Stdin.Fd())
	// Перенос Stdin в non-block для корректного выхода по контексту
	flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
	if err != nil {
		cancel(err)
		return
	}
	_, err = unix.FcntlInt(uintptr(fd), unix.F_SETFL, flags|unix.O_NONBLOCK)
	if err != nil {
		cancel(err)
		return
	}

	defer func() {
		_, err = unix.FcntlInt(uintptr(fd), unix.F_SETFL, flags)
		if err != nil {
			cancel(err)
		}
	}()

	// Восстановление ISIG после Raw Mode
	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		cancel(err)
		return
	}
	termios.Lflag |= unix.ISIG
	if err := unix.IoctlSetTermios(fd, unix.TCSETS, termios); err != nil {
		cancel(err)
		return
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	for {
		select {
		case <-ctx.Done():
			return
		case <-sigChan:
			cancel(nil)
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
		if buf[0] == '\n' || buf[0] == '\r' {
			return enterKeyCode, nil
		} else if buf[0] == 'q' {
			return quitKeyCode, nil
		} else if buf[0] == 'k' {
			return upKeyCode, nil
		} else if buf[0] == 'j' {
			return downKeyCode, nil
		} else if buf[0] == 'l' {
			return enterKeyCode, nil
		}
	}
	// Обрабатывает 'й' как 'q'
	if n == 2 {
		r, _ := utf8.DecodeRune(buf[:])
		if r == 'й' {
			return quitKeyCode, nil
		} else if r == 'л' {
			return upKeyCode, nil
		} else if r == 'о' {
			return downKeyCode, nil
		} else if r == 'д' {
			return enterKeyCode, nil
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
