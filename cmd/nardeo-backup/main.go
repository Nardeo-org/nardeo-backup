package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nardeo/nardeo-backup/internal/ui"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := newRootCommand().ExecuteContext(ctx); err != nil {
		ui.NewLogger(os.Stderr, false).Error("Command failed", "error", err)
		stop()
		os.Exit(1)
	}
}
