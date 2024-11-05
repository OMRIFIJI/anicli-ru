package cli

const (
    QuitCode = -1
    EnterCode = 0
    continueCode = 1
)

const (
    cursorScrollOffset = 2

    cursorStateHigh = 1
    cursorStateNormal = 0
    cursorStateLow = -1
)

const (
	colorReset      = "\033[0m"
	colorPrompt     = "\033[34m"
	colorErr        = "\033[31m"
	highlightBg     = "\033[48;5;235m"
	highlightFg     = "\033[37m"
	highlightCursor = "\033[34m"
	highlightBgReset  = "\033[0m"
)
