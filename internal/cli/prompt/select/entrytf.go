package promptselect

import (
	"strings"
)

const (
	borderStartInactive string = "│ " + highlightBg + " " + highlightBgReset
	borderEndInactive   string = " │"
)

type indexOptions struct {
	index     int
	showIndex bool
}

func fitEntryLines(entry string, termWidth int, indOpts indexOptions) fittedEntry {
	// Сколько текста вмещается в табличку после декорации
	entryLineWidth := termWidth - 7
	entryRune := []rune(entry)
	entryRuneLen := len(entryRune)

	indexPadding := calculatePadding(indOpts)

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
	step := entryLineWidth - indexPadding
	left := entryLineWidth
	for right := entryLineWidth + step; right < entryRuneLen; left, right = left+step, right+step {
		formatAndAppend(string(entryRune[left:right]), 2+indexPadding, 0)
	}

	// Последняя строка, надо заполнить пробелами до конца
	extraSpaces := entryLineWidth - (entryRuneLen - left) - indexPadding
	formatAndAppend(string(entryRune[left:]), 2+indexPadding, extraSpaces)

	return entryStrings
}

func calculatePadding(opts indexOptions) int {
	if !opts.showIndex {
		return 0
	}
	return charLenOfInt(opts.index+1) + 1
}

func fitEntryLine(entryLine string, opts fmtOpts) string {
	// Перенос пробелов с начала строки в конец
	trimmedLine := strings.TrimLeft(entryLine, " ")
	movedSpaces := len(entryLine) - len(trimmedLine)

	var b strings.Builder
	// Длина всей строки, которая будет построена strings.Builder
	totalLen := len(trimmedLine) + len(borderStartInactive) + opts.LeftPadding + movedSpaces + opts.extraSpaces + len(borderEndInactive)
	b.Grow(totalLen)

	b.WriteString(borderStartInactive)
	b.WriteString(strings.Repeat(" ", opts.LeftPadding))
	b.WriteString(trimmedLine)
	b.WriteString(strings.Repeat(" ", movedSpaces+opts.extraSpaces))
	b.WriteString(borderEndInactive)

	return b.String()
}

func makeEntryActive(entry fittedEntry) fittedEntry {
	entryLinesActive := make(fittedEntry, 0, len(entry))

	const borderStartActive string = "│ " + highlightBg + highlightCursor + "▌" + highlightFg
	const borderEndActive string = highlightBgReset + " │"

	var b strings.Builder
	for _, entryStr := range entry {
		// Вырезаем старые borderStartInactive и borderEndInactive
		entryStrWithoutBorder := entryStr[len(borderStartInactive) : len(entryStr)-(len(borderEndInactive))]

		totalLen := len(borderStartActive) + len(entryStrWithoutBorder) + len(borderEndActive)
		b.Grow(totalLen)

		b.WriteString(borderStartActive)
		b.WriteString(entryStrWithoutBorder)
		b.WriteString(borderEndActive)

		entryLinesActive = append(entryLinesActive, b.String())
		b.Reset()
	}

	return entryLinesActive
}
