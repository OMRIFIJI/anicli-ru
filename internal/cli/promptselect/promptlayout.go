package promptselect

import (
	txtclr "anicliru/internal/cli/textcolors"
	"fmt"
	"strconv"
	"strings"
)

type terminalSize struct {
	width  int
	height int
}

type Drawer struct {
    promptMessage string
	entriesLines [][]string
	drawHigh     int
	termSize     terminalSize
	cur          *Cursor
	curVirt      Cursor
}

func (d *Drawer) initInterface(entryList []string, termSize terminalSize, promptMessage string) {
	d.termSize = termSize
    d.promptMessage = promptMessage

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
		posMax: entryCount,
	}
}

func (d *Drawer) drawInterface() {
	if !d.hasCursorMoved() {
		return
	}

	clearScreen()
	hideCursor()

	fmt.Printf("%s%s%s", txtclr.ColorPrompt, d.promptMessage, txtclr.ColorReset)
    moveCursorToNewLine()

	entryCountStr := strconv.Itoa(len(d.entriesLines))
	repeatLineStr := strings.Repeat("─", d.termSize.width-16-len(entryCountStr))
	fmt.Printf("┌───── Всего: %s %s┐", entryCountStr, repeatLineStr)
    moveCursorToNewLine()

	d.drawEntries()
    
	fmt.Printf("└%s┘", strings.Repeat("─", d.termSize.width-2))
}

func (d *Drawer) hasCursorMoved() bool {
	curChange := d.cur.Pos - d.cur.posOld
	// Такой вызов бывает только из начального состояния курсора
	if curChange == 0 {
		return true
	}
	if d.cur.posOld == 0 && curChange < 0 {
		return false
	}
	if d.cur.posOld == d.cur.posMax && curChange > 0 {
		return false
	}
	return true
}

func (d *Drawer) updateDrawParams() {
	curChange := d.cur.Pos - d.cur.posOld
	virtPosNew := d.curVirt.Pos + curChange

	// Если курсор пытается убежать, ничего не меняется
	if virtPosNew > d.curVirt.posMax || virtPosNew < 0 {
		return
	}

	// Не двигаю виртуальный курсор, если добрался до оффсета
	if virtPosNew >= d.curVirt.posMax-(cursorScrollOffset+1) {
		// Настоящий курсор подходит к концу
		if d.cur.Pos >= d.cur.posMax-cursorScrollOffset {
			d.curVirt.Pos = virtPosNew
		} else {
			// Сдвиг entries вниз
			d.drawHigh += curChange
		}
	} else if virtPosNew <= cursorScrollOffset-1 {
		// Настоящий курсор подходит к началу
		if d.cur.Pos <= cursorScrollOffset-1 {
			d.curVirt.Pos = virtPosNew
		} else {
			// Сдвиг entries вверх
			d.drawHigh += curChange
		}
	} else {
		d.curVirt.Pos = virtPosNew
	}
}

func (d *Drawer) drawEntries() {
	d.updateDrawParams()

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
	// Сколько чистого текста вмещается в табличку
	altScreenWidth := termWidth - 7

	// Записываем весь entry в одну строку, если можем
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
