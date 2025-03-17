//go:build windows

package promptselect

import (
	"github.com/OMRIFIJI/anicli-ru/internal/logger"
	"context"
	"os"
	"time"

	"golang.org/x/term"
)

const pollingDelay = 10

func newResizeChan(ctx context.Context) chan struct{} {
	signalChan := make(chan struct{})
	go monitorWindowSize(ctx, signalChan)

	return signalChan
}

func monitorWindowSize(ctx context.Context, signalChan chan struct{}) {
	fd := int(os.Stdout.Fd())
	oldHeight, oldWidth, err := term.GetSize(fd)
	if err != nil {
		logger.ErrorLog.Println(err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollingDelay * time.Millisecond):
			newHeight, newWidth, err := term.GetSize(fd)
			if err != nil {
				logger.ErrorLog.Println(err)
				return
			}

			if newHeight != oldHeight {
				oldHeight = newHeight
				signalChan <- struct{}{}
			}

			if newWidth != oldWidth {
				oldWidth = newWidth
				signalChan <- struct{}{}
			}
		}
	}
}
