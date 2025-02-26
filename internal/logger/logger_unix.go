//go:build !windows

package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	WarnLog  *log.Logger
	ErrorLog *log.Logger
)

func Init() error {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if len(stateHome) == 0 {
		home := os.Getenv("HOME")
		if len(home) == 0 {
			return fmt.Errorf("не удалось создать лог, HOME и XDG_STATE_HOME не заданы.")
		}
		stateHome = home + "/.local/state"
	}
	logPath := stateHome + "/anicli-ru/log.txt"

	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию для лога. %s", err)
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("не удалось открыть лог. %s", err)
	}

	WarnLog = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}
