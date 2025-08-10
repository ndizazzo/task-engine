package testing

import (
	"io"
	"log/slog"
)

// NewDiscardLogger creates a new logger that discards all output
// This is useful for tests to prevent log output from cluttering test results
func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
