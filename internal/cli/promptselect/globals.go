package promptselect

type keyCode uint8
const (
	quitKeyCode keyCode = iota
	enterKeyCode
	continueKeyCode
	upKeyCode
	downKeyCode
)

type exitPromptCode uint8
const (
    onQuitExitCode exitPromptCode = iota
    onEnterExitCode
)

type cursorMoveCode uint8
const (
    upCursorMoveCode cursorMoveCode = iota
    downCursorMoveCode
    noChangeCursorMoveCode
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
