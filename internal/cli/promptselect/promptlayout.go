package promptselect

import (
	txtclr "anicliru/internal/cli/textcolors"
	"fmt"
	"golang.org/x/term"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func (d *Drawer) newDrawer(promptCtx promptContext) {
	d.promptCtx = promptCtx

	d.drawCtx = drawingContext{
		drawHigh:   0,
		virtCurPos: 0,
	}
	d.updateTerminalSize()
	d.fitEntries()

}

func (d *Drawer) fitEntries() {
	d.fittedEntries = nil
	globalLinesCount := 0
	for _, entry := range d.promptCtx.entries {
		fitEntry := fittedEntry{
			lines:     fitEntryLines(entry, d.drawCtx.termSize.width),
			globalInd: globalLinesCount,
		}
		d.fittedEntries = append(d.fittedEntries, fitEntry)
		globalLinesCount += len(fitEntry.lines)
	}
}

func (d *Drawer) spinDrawInterface(keyCodeChan chan keyCode, oldTermState *term.State) {
	defer d.promptCtx.wg.Done()
	defer d.restoreTerm(oldTermState)

	// первый отрисовка интерфейса до нажатия клавиш
	d.drawInterface(noActionKeyCode, false)

	d.promptCtx.wg.Add(1)
    
    quitChan := make(chan bool)
	go d.redrawOnTerminalResize(quitChan, oldTermState)
	defer d.promptCtx.wg.Done()

	for {
		keyCodeValue := <-keyCodeChan
		switch keyCodeValue {
		case upKeyCode, downKeyCode:
			d.drawInterface(keyCodeValue, false)
		case enterKeyCode, quitKeyCode:
			return
		}
	}

}

func (d *Drawer) drawInterface(keyCodeValue keyCode, onResize bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.updateDrawParams(keyCodeValue, onResize)

	clearScreen()

	fmt.Printf("%s%s%s", txtclr.ColorPrompt, d.promptCtx.promptMessage, txtclr.ColorReset)
	moveCursorToNewLine()

	entryCountStr := strconv.Itoa(len(d.fittedEntries))
	repeatLineStr := strings.Repeat("─", d.drawCtx.termSize.width-16-len(entryCountStr))
	fmt.Printf("┌───── Всего: %s %s┐", entryCountStr, repeatLineStr)
	moveCursorToNewLine()

	d.drawEntries()

	fmt.Printf("└%s┘", strings.Repeat("─", d.drawCtx.termSize.width-2))
}

func (d *Drawer) redrawOnTerminalResize(quitChan chan bool, oldTermState *term.State) {
	defer d.promptCtx.wg.Done()
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
	d.drawCtx.termSize = terminalSize{
		width:  termWidth,
		height: termHeight,
	}
}

func (d *Drawer) updateDrawParams(keyCodeValue keyCode, onResize bool) {
	if onResize {
		d.updateTerminalSize()
		d.fitEntries()
        return
	}

	if keyCodeValue == upKeyCode {
		if d.promptCtx.cur.pos == 0 {
			d.drawCtx.virtCurPos = 0
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
		} else if d.promptCtx.cur.pos > len(d.fittedEntries)-1-cursorScrollOffset {
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
	lineCount := 0

	for _, entry := range d.fittedEntries[d.drawCtx.drawHigh:d.promptCtx.cur.pos] {
		for _, line := range entry.lines {
			fmt.Print(line)
			moveCursorToNewLine()
			lineCount++
		}
	}

	selectedEntry := makeEntryActive(d.fittedEntries[d.promptCtx.cur.pos])
	for _, line := range selectedEntry.lines {
		fmt.Print(line)
		moveCursorToNewLine()
		lineCount++
		if lineCount >= d.drawCtx.termSize.height-3 {
			d.drawCtx.drawLow = d.promptCtx.cur.pos
			return
		}
	}

	for i, entry := range d.fittedEntries[d.promptCtx.cur.pos+1:] {
		for _, line := range entry.lines {
			fmt.Print(line)
			moveCursorToNewLine()
			lineCount++
			if lineCount >= d.drawCtx.termSize.height-3 {
				d.drawCtx.drawLow = d.promptCtx.cur.pos + 1 + i
				return
			}
		}
	}

	d.drawCtx.drawLow = len(d.fittedEntries) - 1
}

func (d *Drawer) restoreTerm(oldTermState *term.State) {
	exitAltScreenBuf()
	term.Restore(0, oldTermState)
	showCursor()
}
