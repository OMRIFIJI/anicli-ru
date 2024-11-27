package promptselect

const (
	quitKeyCode keyCode = iota
	enterKeyCode
	noActionKeyCode
	upKeyCode
	downKeyCode
)

const (
    onQuitExitCode exitPromptCode = iota
    onEnterExitCode
    onErrorExitCode
)

const (
	cursorScrollOffset = 2
	cursorStateHigh   = 1
	cursorStateNormal = 0
	cursorStateLow    = -1
)

const (
	highlightBg      = "\033[48;5;235m"
	highlightFg      = "\033[37m"
	highlightCursor  = "\033[34m"
	highlightBgReset = "\033[0m"
)

const (
    resizeDebounceMs = 100
)

const (
    minimalTermHeight = 4
    decorateTextWidth = 16
    borderSize = 1
)
