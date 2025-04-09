package promptselect

import (
	"sync"
)

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
	ch        promptChannels
}

type promptChannels struct {
	keyCode  chan keyCode
	exitCode chan exitPromptCode
}

type promptContext struct {
	promptMessage string
	entries       []string
	cur           int
}

type terminalSize struct {
	width  int
	height int
}

type drawer struct {
	promptCtx promptContext
	drawCtx   drawingContext
	mutex     sync.Mutex
}

type drawingContext struct {
	showIndex     bool
	fittedEntries []fittedEntry
	fittedPrompt  string
	drawHigh      int // Индекс первого entry, видимого на экране
	drawLow       int // Индекс последнего entry, видимого на экране	
	virtCur       int
	termSize      terminalSize
}
