package logger

import (
	"log"
	"os"
	"path/filepath"
)

var (
	WarnLog  *log.Logger
	ErrorLog *log.Logger
)

func Init() {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if len(stateHome) == 0 {
		home := os.Getenv("HOME")
		if len(home) == 0 {
			return
		}
		stateHome = home + "/.local/state"
	}
	logPath := stateHome + "/anicli-ru/log.txt"

	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return
	}

	WarnLog = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
