package promptselect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/OMRIFIJI/anicli-ru/internal/cli/ansi"
	"golang.org/x/term"
)

func newDrawer(promptCtx promptContext, showIndex bool) (*drawer, error) {
	d := &drawer{
		promptCtx: promptCtx,
		drawCtx: drawingContext{
			drawHigh:  0,
			virtCur:   0,
			showIndex: showIndex,
		},
	}

	if showIndex {
		for i := range d.promptCtx.entries {
			d.promptCtx.entries[i] = fmt.Sprintf("%d %s", i+1, d.promptCtx.entries[i])
		}
	}

	if err := d.updateTerminalSize(); err != nil {
		return nil, err
	}
	d.fitPrompt()
	d.fitEntries()

	return d, nil
}

func (d *drawer) fitEntries() {
	d.drawCtx.fittedEntries = make([]fittedEntry, 0, len(d.promptCtx.entries))
	indOpt := indexOptions{showIndex: d.drawCtx.showIndex}

	for i, entry := range d.promptCtx.entries {
		indOpt.index = i
		d.drawCtx.fittedEntries = append(d.drawCtx.fittedEntries,
			fitEntryLines(entry, d.drawCtx.termSize.width, indOpt))
	}
}

func (d *drawer) fitPrompt() {
	runePrompt := []rune(d.promptCtx.promptMessage)
	maxWidth := d.drawCtx.termSize.width - 2*borderSize

	const ellipsis string = "..."

	if len(runePrompt) > maxWidth {
		d.drawCtx.fittedPrompt = string(runePrompt[:maxWidth-len(ellipsis)]) + ellipsis
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
		d.handleResizeEvents(ctx, cancel)
	}()

	defer func() {
		d.recoverWithCancel(cancel)
		wg.Wait()
	}()

	for {
		select {
		case keyCodeValue := <-keyCodeChan:
			if err := d.handleKeyInput(keyCodeValue); err != nil {
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
		return d.drawInterface(keyCodeValue, false)
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

	fmt.Print(d.buildInterfaceStr())

	return nil
}

// Возвращает строку, в которой хранится весь интерфейс prompt select.
func (d *drawer) buildInterfaceStr() string {
	var b strings.Builder

	entryCount := len(d.drawCtx.fittedEntries)
	repeatLineStr := strings.Repeat("─", d.drawCtx.termSize.width-decorateTextWidth-charLenOfInt(entryCount))

	b.WriteString(ansi.ClearScreen)
	fmt.Fprintf(&b, "%s%s%s\r\n", ansi.ColorPrompt, d.drawCtx.fittedPrompt, ansi.ColorReset)
	fmt.Fprintf(&b, "┌───── Всего: %d %s┐\r\n", entryCount, repeatLineStr)

	d.buildEntriesBody(&b)
	fmt.Fprintf(&b, "└%s┘", strings.Repeat("─", d.drawCtx.termSize.width-2))
	return b.String()
}

func (d *drawer) handleResizeEvents(ctx context.Context, cancel context.CancelCauseFunc) {
	defer d.recoverWithCancel(cancel)
	signalChan := newResizeChan(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-signalChan:
			d.debounce()
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
	fd := int(os.Stdout.Fd())
	termWidth, termHeight, err := term.GetSize(fd)
	if err != nil {
		return err
	}

	entryCount := len(d.drawCtx.fittedEntries)
	minimalTermWidth := decorateTextWidth + charLenOfInt(entryCount)
	if termWidth < minimalTermWidth || termHeight < minimalTermHeight {
		return fmt.Errorf("терминал слишком маленький! Минимальный размер: (%dx%d)", minimalTermWidth, minimalTermHeight)
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
		d.correctOnRedraw()
		return nil
	}

	if d.drawCtx.drawLow-d.drawCtx.drawHigh <= cursorScrollOffset {
		d.smallWindowKeyHandle(keyCodeValue)
	} else {
		d.bigWindowKeyHandle(keyCodeValue)
	}

	return nil
}

// Корректирует параметры отрисовки, если выбран entry внизу экрана, и высота уменьшена.
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
}

// Рассчитывает drawHigh, гарантируя, что entry с индексом low поместится на экран.
func (d *drawer) correctDrawHigh(low int) {
	linesLeft := d.drawCtx.termSize.height - 3

	// Проверка на граничный случай - помещается только последний entry
	lastEntryLinesCount := len(d.drawCtx.fittedEntries[low])
	if linesLeft <= lastEntryLinesCount {
		d.drawCtx.drawHigh = low
	}

	// Двигает drawHigh с конца вверх пока не захватит все возможное пространство
	lineCount := 0
	for i := low; i >= 0; i-- {
		line := d.drawCtx.fittedEntries[i]
		lineCount += len(line)
		if lineCount > linesLeft {
			d.drawCtx.drawHigh = i + 1
			return
		}
	}
	d.drawCtx.drawHigh = 0
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
			if d.drawCtx.drawLow < len(d.drawCtx.fittedEntries)-1 {
				d.correctDrawHigh(d.promptCtx.cur)
			}
			d.drawCtx.virtCur = d.drawCtx.drawLow - d.drawCtx.drawHigh
		} else if d.drawCtx.virtCur == d.drawCtx.drawLow-d.drawCtx.drawHigh {
			d.correctDrawHigh(d.promptCtx.cur)
			d.drawCtx.virtCur = d.drawCtx.drawHigh - d.promptCtx.cur
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
	// Клавиша вниз - выглядит сложнее, но полная аналогия с клавишей вверх
	// кроме дополнительного условия correctDrawHighAtLowest.
	if keyCodeValue == downKeyCode {
		if d.promptCtx.cur == len(d.drawCtx.fittedEntries)-1 {
			if d.drawCtx.virtCur == d.drawCtx.drawLow-d.drawCtx.drawHigh && d.drawCtx.drawLow != len(d.drawCtx.fittedEntries)-1 {
				d.correctDrawHigh(d.promptCtx.cur)
			}
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

func (d *drawer) buildEntriesBody(b *strings.Builder) {
	linesLeft := d.drawCtx.termSize.height - 3

	// Отрисовка элементов до курсора
	for i := d.drawCtx.drawHigh; i < d.promptCtx.cur; i++ {
		entry := d.drawCtx.fittedEntries[i]
		for _, line := range entry {
			b.WriteString(line)
			linesLeft--
			if linesLeft <= 0 {
				d.drawCtx.drawLow = i
				return
			}
		}
	}

	// Отрисовка выделенного элемента
	selectedEntry := makeEntryActive(d.drawCtx.fittedEntries[d.promptCtx.cur])
	for _, line := range selectedEntry {
		b.WriteString(line)
		linesLeft--
		if linesLeft <= 0 {
			d.drawCtx.drawLow = d.promptCtx.cur
			return
		}
	}

	// Отрисовка элементов после курсора
	for i := d.promptCtx.cur + 1; i < len(d.drawCtx.fittedEntries); i++ {
		entry := d.drawCtx.fittedEntries[i]
		for _, line := range entry {
			b.WriteString(line)
			linesLeft--
			if linesLeft <= 0 {
				d.drawCtx.drawLow = i
				return
			}
		}
	}

	d.drawCtx.drawLow = len(d.drawCtx.fittedEntries) - 1
}
func (d *drawer) recoverWithCancel(cancel context.CancelCauseFunc) {
	if r := recover(); r != nil {
		var err error
		if recoveredErr, ok := r.(error); ok {
			err = recoveredErr
		} else {
			err = errors.New("неизвестная ошибка в графике")
		}

		cancel(err)
	}
}
