package ui

import (
	"io"
	"log/slog"
)

func NewLogger(out io.Writer, verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: level}))
}
