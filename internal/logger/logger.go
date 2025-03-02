package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

var (
	WarnLog  *log.Logger
	ErrorLog *log.Logger
)

func Init() error {
	logHome, err := xdg.StateFile("anicli-ru")
	if err != nil {
		return err
	}
	logPath := filepath.Join(logHome, "log.txt")

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
