package promptselect

const (
	quitCode     = -1
	enterCode    = 0
	continueCode = 1
	cursorCode   = 2
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
