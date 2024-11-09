package promptselect

import (
    "strings"
    "fmt"
)

func fitEntryLines(entry string, termWidth int) []string {
	// Сколько текста вмещается в табличку после декорации
	altScreenWidth := termWidth - 7
	entryRune := []rune(entry)
	entryRuneLen := len(entryRune)

	var entryStrings []string

	formatAndAppend := func(substring string, leftPadding, extraSpaces int) {
		entryStrings = append(entryStrings,
			fitEntryLine(substring, fmtOpts{
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


func fitEntryLine(entryLine string, opts fmtOpts) string {
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

func makeEntryActive(entry fittedEntry) fittedEntry {
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
