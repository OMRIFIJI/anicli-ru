package promptselect

import (
	"anicliru/internal/cli/ansi"
	clilog "anicliru/internal/cli/log"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

func newDrawer(promptCtx promptContext) (*drawer, error) {
	d := &drawer{}
	d.promptCtx = promptCtx

	d.drawCtx = drawingContext{
		drawHigh:   0,
		virtCurPos: 0,
	}

	d.ch = drawerChannels{
		quitSpin:   make(chan bool, 1),
		quitRedraw: make(chan bool, 1),
	}

	if err := d.updateTerminalSize(); err != nil {
		return nil, err
	}
	d.fitPrompt()
	d.fitEntries()

	return d, nil
}

func (d *drawer) fitEntries() {
	d.fittedEntries = nil
	for _, entry := range d.promptCtx.entries {
		fitEntry := fitEntryLines(entry, d.drawCtx.termSize.width)
		d.fittedEntries = append(d.fittedEntries, fitEntry)
	}
}

func (d *drawer) fitPrompt() {
	runePrompt := []rune(d.promptCtx.promptMessage)
	promptLen := len(runePrompt)
	decorateBoxWidth := d.drawCtx.termSize.width - 2*borderSize
	if promptLen > decorateBoxWidth {
		d.fittedPrompt = string(runePrompt[:decorateBoxWidth-3]) + "..."
	} else {
		d.fittedPrompt = d.promptCtx.promptMessage
	}
}

func (d *drawer) spinDrawInterface(keyCodeChan chan keyCode, errChan chan error) {
	defer d.promptCtx.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			var err error
			if recoveredErr, ok := r.(error); ok {
				err = recoveredErr
			} else {
				err = errors.New("Неизвестная ошибка в графике.")
			}
			d.ch.quitRedraw <- true
			errChan <- err
		}
	}()

	// первая отрисовка интерфейса до нажатия клавиш
	if err := d.drawInterface(noActionKeyCode, false); err != nil {
		d.ch.quitRedraw <- true
		errChan <- err
		return
	}

	d.promptCtx.wg.Add(1)
	go d.redrawOnTerminalResize(errChan)

	for {
		select {
		case keyCodeValue := <-keyCodeChan:
			switch keyCodeValue {
			case upKeyCode, downKeyCode:
				if err := d.drawInterface(keyCodeValue, false); err != nil {
					d.ch.quitRedraw <- true
					errChan <- err
					return
				}
			case enterKeyCode, quitKeyCode:
				d.ch.quitRedraw <- true
				return
			}
		case <-d.ch.quitSpin:
			return
		}
	}
}

func (d *drawer) drawInterface(keyCodeValue keyCode, onResize bool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.updateDrawParams(keyCodeValue, onResize); err != nil {
		return err
	}

	ansi.ClearScreen()

	fmt.Printf("%s%s%s", ansi.ColorPrompt, d.fittedPrompt, ansi.ColorReset)
	ansi.MoveCursorToNewLine()

	entryCountStr := strconv.Itoa(len(d.fittedEntries))
	repeatLineStr := strings.Repeat("─", d.drawCtx.termSize.width-decorateTextWidth-len(entryCountStr))
	fmt.Printf("┌───── Всего: %s %s┐", entryCountStr, repeatLineStr)
	ansi.MoveCursorToNewLine()

	d.drawEntries()

	fmt.Printf("└%s┘", strings.Repeat("─", d.drawCtx.termSize.width-2))
	return nil
}

func (d *drawer) redrawOnTerminalResize(errChan chan error) {
	defer d.promptCtx.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			var err error
			if recoveredErr, ok := r.(error); ok {
				err = recoveredErr
			} else {
				err = errors.New("Неизвестная ошибка в графике.")
			}
			d.ch.quitSpin <- true
			errChan <- err
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGWINCH)

	for {
		d.debounce()

		select {
		case <-d.ch.quitRedraw:
			return
		case <-signalChan:
			if err := d.drawInterface(noActionKeyCode, true); err != nil {
				println("Err channel fill")
				d.ch.quitSpin <- true
				errChan <- err
				return
			}
		}
	}
}

func (d *drawer) debounce() {
	time.Sleep(resizeDebounceMs * time.Millisecond)
}

func (d *drawer) updateTerminalSize() error {
	termWidth, termHeight, err := term.GetSize(0)
	if err != nil {
		return err
	}

	entryCountStr := strconv.Itoa(len(d.fittedEntries))
	minimalTermWidth := decorateTextWidth + len(entryCountStr)
	if termWidth < minimalTermWidth || termHeight < minimalTermHeight {
		errorStr := "Терминал слишком маленький!\n"
		errorStr += fmt.Sprintf("Минимальный размер: (%dx%d).", minimalTermWidth, minimalTermHeight)
		return errors.New(errorStr)
	}
	d.drawCtx.termSize = terminalSize{
		width:  termWidth,
		height: termHeight,
	}

	return nil
}

func (d *drawer) updateDrawParams(keyCodeValue keyCode, onResize bool) error {
	if onResize {
		if err := d.updateTerminalSize(); err != nil {
			return err
		}
		d.fitPrompt()
		d.fitEntries()
		return nil
	}

	if keyCodeValue == upKeyCode {
		if d.promptCtx.cur.pos == 0 {
			d.drawCtx.virtCurPos = 0
		} else if d.drawCtx.drawLow-d.drawCtx.drawHigh <= cursorScrollOffset { // случай маленького окна
			if d.promptCtx.cur.pos < cursorScrollOffset {
				d.drawCtx.virtCurPos--
			}
			d.drawCtx.drawHigh--
		} else if d.promptCtx.cur.pos < cursorScrollOffset {
			d.drawCtx.virtCurPos--
		} else if d.drawCtx.drawHigh > 0 && d.drawCtx.virtCurPos <= cursorScrollOffset {
			d.drawCtx.drawHigh--
		} else {
			d.drawCtx.virtCurPos--
		}
	}
	// Клавиша вниз - сложнее, но полная аналогия с клавишей вверх
	if keyCodeValue == downKeyCode {
		if d.promptCtx.cur.pos == len(d.fittedEntries)-1 {
			d.drawCtx.virtCurPos = d.drawCtx.drawLow - d.drawCtx.drawHigh
		} else if d.drawCtx.drawLow-d.drawCtx.drawHigh <= cursorScrollOffset {
			if d.promptCtx.cur.pos > len(d.fittedEntries)-1-cursorScrollOffset {
				d.drawCtx.virtCurPos++
			}
			d.drawCtx.drawHigh++
		} else if d.promptCtx.cur.pos > len(d.fittedEntries)-1-cursorScrollOffset {
			d.drawCtx.virtCurPos++
		} else if d.drawCtx.drawLow < len(d.fittedEntries)-1 &&
			d.drawCtx.virtCurPos >= d.drawCtx.drawLow-d.drawCtx.drawHigh-cursorScrollOffset {
			d.drawCtx.drawHigh++
		} else {
			d.drawCtx.virtCurPos++
		}
	}

	if d.drawCtx.virtCurPos < 0 {
		clilog.ErrorLog.Println("Virt cur pos error!")
	}
	if d.drawCtx.drawHigh < 0 {
		clilog.ErrorLog.Println("Draw high error!")
	}

	return nil
}

func (d *drawer) drawEntries() {
	lineCount := 0

	for _, entry := range d.fittedEntries[d.drawCtx.drawHigh:d.promptCtx.cur.pos] {
		for _, line := range entry {
			fmt.Print(line)
			ansi.MoveCursorToNewLine()
			lineCount++
		}
	}

	selectedEntry := makeEntryActive(d.fittedEntries[d.promptCtx.cur.pos])
	for _, line := range selectedEntry {
		fmt.Print(line)
		ansi.MoveCursorToNewLine()
		lineCount++
		if lineCount >= d.drawCtx.termSize.height-3 {
			d.drawCtx.drawLow = d.promptCtx.cur.pos
			return
		}
	}

	for i, entry := range d.fittedEntries[d.promptCtx.cur.pos+1:] {
		for _, line := range entry {
			fmt.Print(line)
			ansi.MoveCursorToNewLine()
			lineCount++
			if lineCount >= d.drawCtx.termSize.height-3 {
				d.drawCtx.drawLow = d.promptCtx.cur.pos + 1 + i
				return
			}
		}
	}

	d.drawCtx.drawLow = len(d.fittedEntries) - 1
}
