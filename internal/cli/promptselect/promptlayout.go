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
)

type terminalSize struct {
	width  int
	height int
}

type Drawer struct {
	promptMessage string
	entriesLines  [][]string
	drawHigh      int
	termSize      terminalSize
	cur           *Cursor
	curVirt       Cursor
}

func (d *Drawer) newDrawer(entryList []string, termSize terminalSize, promptMessage string, cur *Cursor) {
	d.termSize = termSize
	d.promptMessage = promptMessage
	d.cur = cur

	for _, entry := range entryList {
		d.entriesLines = append(d.entriesLines, d.fitEntryLines(entry, d.termSize.width))
	}

	lineCount := 0
	entryCount := 0
	for _, entry := range d.entriesLines {
		entryCount++
		lineCount += len(entry)
		if lineCount >= d.termSize.height-3 {
			break
		}
	}

	d.drawHigh = 0
	d.curVirt = Cursor{
		Pos:    0,
		posMax: entryCount - 1,
	}
}

func (d *Drawer) spinDrawInterface(keyCodeChan chan keyCode) {
	enterAltScreenBuf()
	defer exitAltScreenBuf()

	oldTermState, err := term.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer term.Restore(0, oldTermState)

	hideCursor()
	defer showCursor()

	quitRedrawOnResizeChan := make(chan bool, 1)
	go d.redrawOnTerminalResize(quitRedrawOnResizeChan, oldTermState)

	// первый отрисовка интерфейса до нажатия клавиш
	d.drawInterface(noChangeCursorMoveCode)

	for {
		keyCodeValue := <-keyCodeChan
		switch keyCodeValue {
		case upKeyCode:
			d.drawInterface(upCursorMoveCode)
		case downKeyCode:
			d.drawInterface(downCursorMoveCode)
		case enterKeyCode, quitKeyCode:
			quitRedrawOnResizeChan <- true
			return
		}
	}

}

func (d *Drawer) drawInterface(moveCode cursorMoveCode) {
	clearScreen()

	fmt.Printf("%s%s%s", txtclr.ColorPrompt, d.promptMessage, txtclr.ColorReset)
	moveCursorToNewLine()

	entryCountStr := strconv.Itoa(len(d.entriesLines))
	repeatLineStr := strings.Repeat("─", d.termSize.width-16-len(entryCountStr))
	fmt.Printf("┌───── Всего: %s %s┐", entryCountStr, repeatLineStr)
	moveCursorToNewLine()

	d.drawEntries(moveCode)

	fmt.Printf("└%s┘", strings.Repeat("─", d.termSize.width-2))

}

func (d *Drawer) redrawOnTerminalResize(quitChan chan bool, oldTermState *term.State) {
    defer func(){
        if r := recover(); r != nil {
	        exitAltScreenBuf()
	        term.Restore(0, oldTermState)
            showCursor()
        }
    }()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGWINCH)

	for {
		<-signalChan
		select {
		case <-quitChan:
			return
		default:
			d.drawInterface(noChangeCursorMoveCode)
		}
	}
}

func (d *Drawer) updateDrawParams(moveCode cursorMoveCode) {
	virtPosNew := d.curVirt.Pos

	switch moveCode {
	case upCursorMoveCode:
		virtPosNew--
	case downCursorMoveCode:
		virtPosNew++
	case noChangeCursorMoveCode:
		return
	}

	// Если курсор пытается убежать, ничего не меняется
	if virtPosNew > d.curVirt.posMax || virtPosNew < 0 {
		return
	}

	// Не двигаю виртуальный курсор, если добрался до оффсета
	if virtPosNew >= d.curVirt.posMax-cursorScrollOffset {
		// Настоящий курсор подходит к концу
		if d.cur.Pos >= d.cur.posMax-cursorScrollOffset {
			d.curVirt.Pos = virtPosNew
		} else {
			// Сдвиг entries вниз
			d.drawHigh++
		}
	} else if virtPosNew <= cursorScrollOffset-1 {
		// Настоящий курсор подходит к началу
		if d.cur.Pos <= cursorScrollOffset-1 {
			d.curVirt.Pos = virtPosNew
		} else {
			// Сдвиг entries вверх
			d.drawHigh--
		}
	} else {
		d.curVirt.Pos = virtPosNew
	}
}

func (d *Drawer) drawEntries(moveCode cursorMoveCode) {
	d.updateDrawParams(moveCode)

	// Нужно отдельно обработать строку с курсором...
	lineCount := 0

	for _, entry := range d.entriesLines[d.drawHigh:d.cur.Pos] {
		for _, line := range entry {
			fmt.Print(line)
			moveCursorToNewLine()
			lineCount++
		}
	}

	for _, line := range d.makeEntryActive(d.entriesLines[d.cur.Pos]) {
		fmt.Print(line)
		moveCursorToNewLine()
		lineCount++
		if lineCount >= d.termSize.height-4 {
			return
		}
	}

	if d.cur.Pos+1 >= len(d.entriesLines) {
		return
	}

	for _, entry := range d.entriesLines[d.cur.Pos+1:] {
		for _, line := range entry {
			fmt.Print(line)
			moveCursorToNewLine()
			lineCount++
			if lineCount >= d.termSize.height-4 {
				return
			}
		}
	}
}

type fmtOpts struct {
	extraSpaces int
	LeftPadding int
}

func (d *Drawer) fitEntryLines(entry string, termWidth int) []string {
	var entryStrings []string
	entryRune := []rune(entry)
	entryRuneLen := len(entryRune)
	// Столько чистого текста вмещается в табличку
	altScreenWidth := termWidth - 7

	// Записываем весь entry в одну строку, если можем.
	// Если можем, то надо заполнять пробелами до конца
	if entryRuneLen <= altScreenWidth {
		extraSpaces := altScreenWidth - entryRuneLen
		entryStrings = append(
			entryStrings,
			d.formatLine(
				string(entryRune[:entryRuneLen]),
				fmtOpts{
					extraSpaces: extraSpaces,
					LeftPadding: 2,
				},
			),
		)
		return entryStrings
	} else {
		entryStrings = append(
			entryStrings,
			d.formatLine(
				string(entryRune[:altScreenWidth]),
				fmtOpts{
					extraSpaces: 0,
					LeftPadding: 2,
				},
			),
		)
	}

	// Остальные строки entry кроме последней
	left := altScreenWidth
	right := left + altScreenWidth - 2
	newLinesCount := 1
	for right < entryRuneLen {
		entryStrings = append(entryStrings,
			d.formatLine(
				string(entryRune[left:right]),
				fmtOpts{
					extraSpaces: 0,
					LeftPadding: 4,
				},
			),
		)
		newLinesCount += 1
		left += altScreenWidth - 2
		right += altScreenWidth - 2
	}

	// Последняя строка, надо заполнить пробелами до конца
	extraSpaces := altScreenWidth - 2 - (entryRuneLen - left)
	entryStrings = append(entryStrings,
		d.formatLine(
			string(entryRune[left:]),
			fmtOpts{
				extraSpaces: extraSpaces,
				LeftPadding: 4,
			},
		),
	)

	return entryStrings
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

func (d *Drawer) makeEntryActive(entry []string) []string {
	entryActive := make([]string, 0, len(entry))

	for _, entryStr := range entry {
		var b strings.Builder
		entryRune := []rune(entryStr)
		b.WriteString("│ ")
		fmt.Fprintf(&b, "%s%s▌ %s", highlightBg, highlightCursor, highlightFg)
		// Фокус с подсчётом рун
		b.WriteString(string(entryRune[19 : len(entryRune)-2]))
		b.WriteString(highlightBgReset)
		b.WriteString(" │")
		entryActive = append(entryActive, b.String())
	}
	return entryActive
}
