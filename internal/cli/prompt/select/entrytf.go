package promptselect

import (
	"fmt"
	"strings"
)

func fitEntryLines(entry string, index, termWidth int) fittedEntry {
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
	for right := left + altScreenWidth - 2; right < entryRuneLen; left, right = left+altScreenWidth-2, right+altScreenWidth-2 {
		formatAndAppend(string(entryRune[left:right]), 3+charLenOfInt(index+1), 0)
	}

	// Последняя строка, надо заполнить пробелами до конца
	extraSpaces := altScreenWidth - (entryRuneLen - left) - 1 - charLenOfInt(index+1)
	formatAndAppend(string(entryRune[left:]), 3+charLenOfInt(index+1), extraSpaces)

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
	entryLinesActive := make([]string, 0, len(entry))

	for _, entryStr := range entry {
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

	return entryLinesActive
}
