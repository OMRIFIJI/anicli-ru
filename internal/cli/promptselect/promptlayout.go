package promptselect

import (
	txtclr "anicliru/internal/cli/textcolors"
	"fmt"
	"golang.org/x/term"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type terminalSize struct {
	width  int
	height int
}

type drawingContext struct {
	drawHigh            int // Индекс самого первого entry видимого на экране
	drawLow             int // Аналогично
	displayedLinesCount int
	virtCurPos          int
}

type fittedEntry struct {
	lines     []string
	globalInd int
}

type Drawer struct {
	promptMessage string
	entries       []string
	fittedEntries []fittedEntry
	termSize      terminalSize
	cur           *Cursor
	drawCtx       drawingContext
	mutex         sync.Mutex
}

func (d *Drawer) newDrawer(entries []string, promptMessage string, cur *Cursor) {
	d.promptMessage = promptMessage
	d.entries = entries
	d.cur = cur

	d.updateTerminalSize()
	d.fitEntries()

	d.drawCtx = drawingContext{
		drawHigh:   0,
		virtCurPos: 0,
	}
}

func (d *Drawer) fitEntries() {
	d.fittedEntries = nil
	globalLinesCount := 0
	for _, entry := range d.entries {
		fitEntry := fittedEntry{
			lines:     d.fitEntryLines(entry, d.termSize.width),
			globalInd: globalLinesCount,
		}
		d.fittedEntries = append(d.fittedEntries, fitEntry)
		globalLinesCount += len(fitEntry.lines)
	}
}

func (d *Drawer) spinDrawInterface(keyCodeChan chan keyCode, oldTermState *term.State, wg *sync.WaitGroup) {
	defer wg.Done()
	defer d.restoreTerm(oldTermState)

	// первый отрисовка интерфейса до нажатия клавиш
	d.drawInterface(noActionKeyCode, false)

	wg.Add(1)
	quitRedrawOnResizeChan := make(chan bool, 1)
	go d.redrawOnTerminalResize(quitRedrawOnResizeChan, oldTermState, wg)

	for {
		keyCodeValue := <-keyCodeChan
		switch keyCodeValue {
		case upKeyCode, downKeyCode:
			d.drawInterface(keyCodeValue, false)
		case enterKeyCode, quitKeyCode:
			quitRedrawOnResizeChan <- true
			return
		}
	}

}

func (d *Drawer) drawInterface(keyCodeValue keyCode, onResize bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.updateDrawParams(keyCodeValue, onResize)

	clearScreen()

	fmt.Printf("%s%s%s", txtclr.ColorPrompt, d.promptMessage, txtclr.ColorReset)
	moveCursorToNewLine()

	entryCountStr := strconv.Itoa(len(d.fittedEntries))
	repeatLineStr := strings.Repeat("─", d.termSize.width-16-len(entryCountStr))
	fmt.Printf("┌───── Всего: %s %s┐", entryCountStr, repeatLineStr)
	moveCursorToNewLine()

	d.drawEntries()

	fmt.Printf("└%s┘", strings.Repeat("─", d.termSize.width-2))
}

func (d *Drawer) redrawOnTerminalResize(quitChan chan bool, oldTermState *term.State, wg *sync.WaitGroup) {
	defer wg.Done()
	defer d.restoreTerm(oldTermState)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGWINCH)

	for {
		d.debounce()

		select {
		case <-quitChan:
			return
		case <-signalChan:
			d.drawInterface(noActionKeyCode, true)
		}
	}
}

func (d *Drawer) debounce() {
	time.Sleep(resizeDebounceMs * time.Millisecond)
}

func (d *Drawer) updateTerminalSize() {
	termWidth, termHeight, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}
	d.termSize = terminalSize{
		width:  termWidth,
		height: termHeight,
	}
}

func (d *Drawer) updateDrawParams(keyCodeValue keyCode, onResize bool) {
	if onResize {
		d.updateTerminalSize()
		d.fitEntries()

	}

	if keyCodeValue == upKeyCode {
		if d.cur.Pos == 0 {
			d.drawCtx.virtCurPos = 0
		} else if d.cur.Pos < cursorScrollOffset {
			d.drawCtx.virtCurPos--
		} else if d.drawCtx.drawHigh > 0 && d.drawCtx.virtCurPos <= cursorScrollOffset {
			d.drawCtx.drawHigh--
		} else {
			d.drawCtx.virtCurPos--
		}
	}
	// Клавиша вниз - сложнее, но полная аналогия с клавишей вверх
	if keyCodeValue == downKeyCode {
		if d.cur.Pos == len(d.fittedEntries)-1 {
			d.drawCtx.virtCurPos = d.drawCtx.drawLow - d.drawCtx.drawHigh
		} else if d.cur.Pos > len(d.fittedEntries)-1-cursorScrollOffset {
			d.drawCtx.virtCurPos++
		} else if d.drawCtx.drawLow < len(d.fittedEntries)-1 &&
			d.drawCtx.virtCurPos >= d.drawCtx.drawLow-d.drawCtx.drawHigh-cursorScrollOffset {
			d.drawCtx.drawHigh++
		} else {
			d.drawCtx.virtCurPos++
		}
	}
}

func (d *Drawer) drawEntries() {
	/*
		fmt.Printf("Virtual cursor pose: %d", d.drawCtx.virtCurPos)
		moveCursorToNewLine()
		fmt.Printf("draw Low: %d", d.drawCtx.drawLow)
		moveCursorToNewLine()
		fmt.Printf("draw High: %d", d.drawCtx.drawHigh)
		moveCursorToNewLine()
	*/

	// Нужно отдельно обработать строку с курсором...
	lineCount := 0

	for _, entry := range d.fittedEntries[d.drawCtx.drawHigh:d.cur.Pos] {
		for _, line := range entry.lines {
			fmt.Print(line)
			moveCursorToNewLine()
			lineCount++
		}
	}

	selectedEntry := d.makeEntryActive(d.fittedEntries[d.cur.Pos])
	for _, line := range selectedEntry.lines {
		fmt.Print(line)
		moveCursorToNewLine()
		lineCount++
		if lineCount >= d.termSize.height-3 {
			d.drawCtx.drawLow = d.cur.Pos
			return
		}
	}

	for i, entry := range d.fittedEntries[d.cur.Pos+1:] {
		for _, line := range entry.lines {
			fmt.Print(line)
			moveCursorToNewLine()
			lineCount++
			if lineCount >= d.termSize.height-3 {
				d.drawCtx.drawLow = d.cur.Pos + 1 + i
				return
			}
		}
	}

	d.drawCtx.drawLow = len(d.fittedEntries) - 1
}

func (d *Drawer) fitEntryLines(entry string, termWidth int) []string {
	// Сколько текста вмещается в табличку после декорации
	altScreenWidth := termWidth - 7
	entryRune := []rune(entry)
	entryRuneLen := len(entryRune)

	var entryStrings []string

	formatAndAppend := func(substring string, leftPadding, extraSpaces int) {
		entryStrings = append(entryStrings,
			d.formatLine(substring, fmtOpts{
				extraSpaces: extraSpaces,
				LeftPadding: leftPadding,
			}),
		)
	}

	// Записываем весь entry в одну строку, если можем
	if entryRuneLen <= altScreenWidth {
		extraSpaces := altScreenWidth - entryRuneLen
		formatAndAppend(string(entryRune), 2, extraSpaces)
		return entryStrings
	}

	// Если не поместилось в одну, записываем первую строку entry
	formatAndAppend(string(entryRune[:altScreenWidth]), 2, 0)

	// Остальные строки entry кроме последней
	left := altScreenWidth
	for right := left + altScreenWidth - 2; right < entryRuneLen; left, right = right+altScreenWidth-2, right+altScreenWidth-2 {
		formatAndAppend(string(entryRune[left:right]), 4, 0)
	}

	// Последняя строка, надо заполнить пробелами до конца
	extraSpaces := altScreenWidth - 2 - (entryRuneLen - left)
	formatAndAppend(string(entryRune[left:]), 4, extraSpaces)

	return entryStrings
}

type fmtOpts struct {
	extraSpaces int
	LeftPadding int
}

func (d *Drawer) formatLine(entryLine string, opts fmtOpts) string {
	var b strings.Builder
	b.WriteString("│ ")

	fmt.Fprintf(&b, "%s %s", highlightBg, highlightBgReset)
	b.WriteString(strings.Repeat(" ", opts.LeftPadding))

	// Перенос пробелов с начала строки в конец
	trimmedLine := strings.TrimLeft(entryLine, " ")
	movedSpaces := len(entryLine) - len(trimmedLine)
	b.WriteString(trimmedLine)
	b.WriteString(strings.Repeat(" ", movedSpaces))

	if opts.extraSpaces > 0 {
		b.WriteString(strings.Repeat(" ", opts.extraSpaces))
	}
	b.WriteString(" │")
	return b.String()
}

func (d *Drawer) makeEntryActive(entry fittedEntry) fittedEntry {
	entryLinesActive := make([]string, 0, len(entry.lines))

	for _, entryStr := range entry.lines {
		var b strings.Builder
		entryRune := []rune(entryStr)
		b.WriteString("│ ")
		fmt.Fprintf(&b, "%s%s▌ %s", highlightBg, highlightCursor, highlightFg)
		// Фокус с подсчётом рун
		b.WriteString(string(entryRune[19 : len(entryRune)-2]))
		b.WriteString(highlightBgReset)
		b.WriteString(" │")
		entryLinesActive = append(entryLinesActive, b.String())
	}

	entryActive := fittedEntry{
		lines:     entryLinesActive,
		globalInd: entry.globalInd,
	}
	return entryActive
}

func (d *Drawer) restoreTerm(oldTermState *term.State) {
	exitAltScreenBuf()
	term.Restore(0, oldTermState)
	showCursor()
}
