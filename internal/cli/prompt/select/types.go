package promptselect

import "sync"

type keyCode uint8
type exitPromptCode uint8
type fittedEntry []string

type fmtOpts struct {
	extraSpaces int
	LeftPadding int
}

type PromptSelect struct {
	promptCtx promptContext
	drawer    *drawer
	termSize  terminalSize
	ch        promptChannels
}

type promptChannels struct {
	keyCode  chan keyCode
	exitCode chan exitPromptCode
	err      chan error
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

type drawer struct {
	fittedEntries []fittedEntry
	fittedPrompt  string
	promptCtx     promptContext
	drawCtx       drawingContext
	mutex         sync.Mutex
	ch            drawerChannels
}

type drawerChannels struct {
	quitSpin   chan bool
	quitRedraw chan bool
}

type drawingContext struct {
	drawHigh            int // Индекс самого первого entry видимого на экране
	drawLow             int // Аналогично
	displayedLinesCount int
	virtCurPos          int
	termSize            terminalSize
}
