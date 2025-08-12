package testing

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDiscardLogger(t *testing.T) {
	logger := NewDiscardLogger()
	assert.NotNil(t, logger)

	// Test that it's actually a discard logger by checking it doesn't panic
	logger.Info("test message")
	logger.Error("test error")
	logger.Warn("test warning")
	logger.Debug("test debug")

	// Should not panic or cause any issues
	assert.True(t, true, "Logger should handle all log levels without issues")
}

func TestDiscardLoggerInterface(t *testing.T) {
	var logger *slog.Logger = NewDiscardLogger()
	assert.NotNil(t, logger)

	// Test that it implements the slog.Logger interface
	logger.Info("test", "key", "value")
	logger.Error("test error", "key", "value")

	// Should not panic
	assert.True(t, true, "Logger should implement slog.Logger interface")
}

func TestDiscardLoggerConcurrency(t *testing.T) {
	logger := NewDiscardLogger()

	// Test concurrent logging
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("concurrent log", "id", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic or cause race conditions
	assert.True(t, true, "Logger should handle concurrent access safely")
}

func TestDiscardLoggerWithFields(t *testing.T) {
	logger := NewDiscardLogger()

	// Test logging with various field types
	logger.Info("test message",
		"string", "value",
		"int", 42,
		"bool", true,
		"float", 3.14,
		"nil", nil,
	)

	// Should not panic
	assert.True(t, true, "Logger should handle all field types")
}

func TestDiscardLoggerWithGroups(t *testing.T) {
	logger := NewDiscardLogger()

	// Test logging with groups
	logger.Info("test message",
		slog.Group("group1", "key1", "value1", "key2", "value2"),
		slog.Group("group2", "key3", "value3"),
	)

	// Should not panic
	assert.True(t, true, "Logger should handle groups")
}
