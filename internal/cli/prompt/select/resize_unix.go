//go:build !windows

package promptselect

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Контекст остался из-за перегрузки для платформ.
// Надо бы разделить.
func newResizeChan(ctx context.Context) chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGWINCH)
	return signalChan
}
