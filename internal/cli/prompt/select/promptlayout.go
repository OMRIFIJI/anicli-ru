package promptselect

import (
	"anicliru/internal/cli/ansi"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"
)

func newDrawer(promptCtx promptContext, showIndex bool) (*drawer, error) {
	d := &drawer{}
	d.promptCtx = promptCtx

	if showIndex {
		for i := 0; i < len(d.promptCtx.entries); i++ {
			d.promptCtx.entries[i] = fmt.Sprintf("%d %s", i+1, d.promptCtx.entries[i])
		}
	}

	d.drawCtx = drawingContext{
		drawHigh:  0,
		virtCur:   0,
		showIndex: showIndex,
	}

	if err := d.updateTerminalSize(); err != nil {
		return nil, err
	}
	d.fitPrompt()
	d.fitEntries()

	return d, nil
}

func (d *drawer) fitEntries() {
	d.drawCtx.fittedEntries = nil
	indOpt := indexOptions{showIndex: d.drawCtx.showIndex}

	for i, entry := range d.promptCtx.entries {
		indOpt.index = i
		fitEntry := fitEntryLines(entry, d.drawCtx.termSize.width, indOpt)
		d.drawCtx.fittedEntries = append(d.drawCtx.fittedEntries, fitEntry)
	}
}

func (d *drawer) fitPrompt() {
	runePrompt := []rune(d.promptCtx.promptMessage)
	promptLen := len(runePrompt)
	decorateBoxWidth := d.drawCtx.termSize.width - 2*borderSize
	if promptLen > decorateBoxWidth {
		d.drawCtx.fittedPrompt = string(runePrompt[:decorateBoxWidth-3]) + "..."
	} else {
		d.drawCtx.fittedPrompt = d.promptCtx.promptMessage
	}
}

func (d *drawer) spinDrawInterface(keyCodeChan chan keyCode, ctx context.Context, cancel context.CancelCauseFunc) {
	// первая отрисовка интерфейса до нажатия клавиш
	if err := d.drawInterface(noActionKeyCode, false); err != nil {
		cancel(err)
		return
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		d.spinRedrawOnResize(ctx, cancel)
	}()
	defer wg.Wait()
	defer d.recoverWithCancel(cancel)

	for {
		select {
		case keyCodeValue := <-keyCodeChan:
			err := d.handleKeyInput(keyCodeValue)
			if err != nil {
				cancel(err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (d *drawer) handleKeyInput(keyCodeValue keyCode) error {
	switch keyCodeValue {
	case upKeyCode, downKeyCode:
		d.moveCursor(keyCodeValue)
		if err := d.drawInterface(keyCodeValue, false); err != nil {
			return err
		}
	}
	return nil
}

func (d *drawer) moveCursor(keyCodeValue keyCode) {
	switch keyCodeValue {
	case downKeyCode:
		if d.promptCtx.cur < len(d.promptCtx.entries)-1 {
			d.promptCtx.cur++
		}
	case upKeyCode:
		if d.promptCtx.cur > 0 {
			d.promptCtx.cur--
		}
	}
}

func (d *drawer) drawInterface(keyCodeValue keyCode, onResize bool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.updateDrawContext(keyCodeValue, onResize); err != nil {
		return err
	}

	entryCount := len(d.drawCtx.fittedEntries)
	repeatLineStr := strings.Repeat("─", d.drawCtx.termSize.width-decorateTextWidth-charLenOfInt(entryCount))

	var drawBuilder strings.Builder

    drawBuilder.WriteString(ansi.ClearScreen)
	fmt.Fprintf(&drawBuilder, "%s%s%s\r\n", ansi.ColorPrompt, d.drawCtx.fittedPrompt, ansi.ColorReset)
	fmt.Fprintf(&drawBuilder, "┌───── Всего: %d %s┐\r\n", entryCount, repeatLineStr)
	d.drawEntriesBody(&drawBuilder)
	fmt.Fprintf(&drawBuilder, "└%s┘", strings.Repeat("─", d.drawCtx.termSize.width-2))

	fmt.Print(drawBuilder.String())

	return nil
}

func (d *drawer) spinRedrawOnResize(ctx context.Context, cancel context.CancelCauseFunc) {
	defer d.recoverWithCancel(cancel)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGWINCH)

	for {
		d.debounce()

		select {
		case <-ctx.Done():
			return
		case <-signalChan:
			if err := d.drawInterface(noActionKeyCode, true); err != nil {
				cancel(err)
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

	entryCount := len(d.drawCtx.fittedEntries)
	minimalTermWidth := decorateTextWidth + charLenOfInt(entryCount)
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

func (d *drawer) updateDrawContext(keyCodeValue keyCode, onResize bool) error {
	if onResize {
		if err := d.updateTerminalSize(); err != nil {
			return err
		}
		d.fitPrompt()
		d.fitEntries()
		// Если выбран нижний вариант, и высота уменьшена
		d.correctOnRedraw()
		return nil
	}

	if d.drawCtx.drawLow-d.drawCtx.drawHigh <= cursorScrollOffset {
		d.smallWindowKeyHandle(keyCodeValue)
	} else {
		d.bigWindowKeyHandle(keyCodeValue)
	}

	// При переключении между маленьким и большим окном курсор может улететь вниз
	d.correctDrawHigh()

	return nil
}

func (d *drawer) correctOnRedraw() {
	newDrawLow := d.drawCtx.drawHigh
	lineCount := 0
	for _, line := range d.drawCtx.fittedEntries[d.drawCtx.drawHigh:] {
		lineCount += len(line)
		if lineCount >= d.drawCtx.termSize.height-3 {
			// Если курсор за пределами экрана
			if newDrawLow-d.drawCtx.drawHigh < d.drawCtx.virtCur {
				d.drawCtx.drawHigh++
				d.drawCtx.virtCur--
			}
			return
		}
		newDrawLow++
	}
	return
}

func (d *drawer) correctDrawHigh() {
	newDrawLow := d.drawCtx.drawHigh
	lineCount := 0
	for _, line := range d.drawCtx.fittedEntries[d.drawCtx.drawHigh:] {
		lineCount += len(line)
		if lineCount >= d.drawCtx.termSize.height-3 {
			// Если курсор за пределами экрана
			if newDrawLow-d.drawCtx.drawHigh < d.drawCtx.virtCur {
				// Сдвигаем вниз, чтобы компенсировать прыжок курсора
				d.drawCtx.drawHigh += d.drawCtx.virtCur - (newDrawLow - d.drawCtx.drawHigh)
			}
			return
		}
		newDrawLow++
	}
	return
}

func (d *drawer) smallWindowKeyHandle(keyCodeValue keyCode) {
	if keyCodeValue == upKeyCode {
		if d.promptCtx.cur == 0 {
			d.drawCtx.virtCur = 0
			d.drawCtx.drawHigh = 0
		} else if d.drawCtx.virtCur == 0 {
			d.drawCtx.drawHigh--
		} else {
			d.drawCtx.virtCur--
		}
	}

	if keyCodeValue == downKeyCode {
		if d.promptCtx.cur == len(d.drawCtx.fittedEntries)-1 {
			d.drawCtx.virtCur = d.drawCtx.drawLow - d.drawCtx.drawHigh
			if d.drawCtx.drawLow < len(d.drawCtx.fittedEntries)-1 {
				d.drawCtx.drawHigh++
			}
		} else if d.drawCtx.virtCur == d.drawCtx.drawLow-d.drawCtx.drawHigh {
			d.drawCtx.drawHigh++
		} else {
			d.drawCtx.virtCur++
		}
	}
}

func (d *drawer) bigWindowKeyHandle(keyCodeValue keyCode) {
	if keyCodeValue == upKeyCode {
		if d.promptCtx.cur == 0 {
			d.drawCtx.virtCur = 0
		} else if d.promptCtx.cur < cursorScrollOffset {
			d.drawCtx.virtCur--
		} else if d.drawCtx.drawHigh > 0 && d.drawCtx.virtCur <= cursorScrollOffset {
			d.drawCtx.drawHigh--
		} else {
			d.drawCtx.virtCur--
		}
	}
	// Клавиша вниз - сложнее, но полная аналогия с клавишей вверх
	if keyCodeValue == downKeyCode {
		if d.promptCtx.cur == len(d.drawCtx.fittedEntries)-1 {
			d.drawCtx.virtCur = d.drawCtx.drawLow - d.drawCtx.drawHigh
		} else if d.promptCtx.cur > len(d.drawCtx.fittedEntries)-1-cursorScrollOffset {
			d.drawCtx.virtCur++
		} else if d.drawCtx.drawLow < len(d.drawCtx.fittedEntries)-1 &&
			d.drawCtx.virtCur >= d.drawCtx.drawLow-d.drawCtx.drawHigh-cursorScrollOffset {
			d.drawCtx.drawHigh++
		} else {
			d.drawCtx.virtCur++
		}
	}

}

func (d *drawer) drawEntriesBody(drawBuilder *strings.Builder) {
	lineCount := 0

	for _, entry := range d.drawCtx.fittedEntries[d.drawCtx.drawHigh:d.promptCtx.cur] {
		for _, line := range entry {
			drawBuilder.WriteString(line)
			lineCount++
		}
	}

	selectedEntry := makeEntryActive(d.drawCtx.fittedEntries[d.promptCtx.cur])
	for _, line := range selectedEntry {
		drawBuilder.WriteString(line)
		lineCount++
		if lineCount >= d.drawCtx.termSize.height-3 {
			d.drawCtx.drawLow = d.promptCtx.cur
			return
		}
	}

	for i, entry := range d.drawCtx.fittedEntries[d.promptCtx.cur+1:] {
		for _, line := range entry {
			drawBuilder.WriteString(line)
			lineCount++
			if lineCount >= d.drawCtx.termSize.height-3 {
				d.drawCtx.drawLow = d.promptCtx.cur + 1 + i
				return
			}
		}
	}

	d.drawCtx.drawLow = len(d.drawCtx.fittedEntries) - 1
	return
}

func (d *drawer) recoverWithCancel(cancel context.CancelCauseFunc) {
	if r := recover(); r != nil {
		var err error
		if recoveredErr, ok := r.(error); ok {
			err = recoveredErr
		} else {
			err = errors.New("Неизвестная ошибка в графике.")
		}

		cancel(err)
	}
}
