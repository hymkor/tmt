package ctrlc

import (
	"context"
	"os"
	"os/signal"
	"time"
)

func Sleep(ctx context.Context, n time.Duration) bool {
	select {
	case <-time.After(n):
		return false
	case <-ctx.Done():
		return true
	}
}

func IsCancel(ctx context.Context) bool {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return true
		default:
		}
	}
	return false
}

func sigint2cancel(sigint chan os.Signal, quit chan struct{}, cancel func()) {
	select {
	case <-sigint:
		cancel()
		return
	case <-quit:
		println("quit")
		return
	}
}

func Setup(_ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(_ctx)

	quit := make(chan struct{}, 1)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	go sigint2cancel(sigint, quit, cancel)

	return ctx, func() { quit <- struct{}{}; close(sigint); close(quit) }
}
