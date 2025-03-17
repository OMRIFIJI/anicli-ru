package promptselect

import (
	"fmt"
	"strings"
)

type indexOptions struct {
	index     int
	showIndex bool
}

func fitEntryLines(entry string, termWidth int, indOpt indexOptions) fittedEntry {
	// Сколько текста вмещается в табличку после декорации
	entryLineWidth := termWidth - 7
	entryRune := []rune(entry)
	entryRuneLen := len(entryRune)

	var indexPadding int
	if indOpt.showIndex {
		indexCharSize := charLenOfInt(indOpt.index + 1)
		indexPadding = indexCharSize + 1
	} else {
		indexPadding = 0
	}

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
	if entryRuneLen <= entryLineWidth {
		extraSpaces := entryLineWidth - entryRuneLen
		formatAndAppend(string(entryRune), 2, extraSpaces)
		return entryStrings
	}

	// Если не поместилось в одну, записываем первую строку entry
	formatAndAppend(string(entryRune[:entryLineWidth]), 2, 0)

	// Остальные строки entry кроме последней
	left := entryLineWidth
	for right := left + entryLineWidth - indexPadding; right < entryRuneLen; left, right = left+entryLineWidth-indexPadding, right+entryLineWidth-indexPadding {
		formatAndAppend(string(entryRune[left:right]), 2+indexPadding, 0)
	}

	// Последняя строка, надо заполнить пробелами до конца
	extraSpaces := entryLineWidth - (entryRuneLen - left) - indexPadding
	formatAndAppend(string(entryRune[left:]), 2+indexPadding, extraSpaces)

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
