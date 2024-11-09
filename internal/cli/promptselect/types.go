package promptselect

import "sync"

type keyCode uint8
type exitPromptCode uint8


type fmtOpts struct {
	extraSpaces int
	LeftPadding int
}

type PromptSelect struct {
	promptCtx     promptContext
	drawer        Drawer
	termSize      terminalSize
}

type promptContext struct {
	promptMessage string
	entries       []string
	cur           *Cursor
	wg            *sync.WaitGroup
}

type Cursor struct {
	pos    int
	posMax int
}

type terminalSize struct {
	width  int
	height int
}


type Drawer struct {
	fittedEntries []fittedEntry
	promptCtx     promptContext
	drawCtx       drawingContext
	mutex         sync.Mutex
}

type drawingContext struct {
	drawHigh            int // Индекс самого первого entry видимого на экране
	drawLow             int // Аналогично
	displayedLinesCount int
	virtCurPos          int
	termSize            terminalSize
}

type fittedEntry struct {
	lines     []string
	globalInd int
}

