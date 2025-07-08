package file

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	task_engine "github.com/ndizazzo/task-engine"
)

type ReplaceLinesAction struct {
	task_engine.BaseAction

	FilePath        string
	ReplacePatterns map[*regexp.Regexp]string
}

func NewReplaceLinesAction(
	filePath string,
	patterns map[*regexp.Regexp]string, logger *slog.Logger,
) *task_engine.Action[*ReplaceLinesAction] {
	return &task_engine.Action[*ReplaceLinesAction]{
		ID: "replace-lines-action",
		Wrapped: &ReplaceLinesAction{
			BaseAction:      task_engine.BaseAction{Logger: logger},
			FilePath:        filePath,
			ReplacePatterns: patterns,
		},
	}
}

func (a *ReplaceLinesAction) Execute(ctx context.Context) error {
	file, err := os.Open(a.FilePath)
	if err != nil {
		a.Logger.Error("Failed to open file",
			"FilePath", a.FilePath,
			"error", err,
		)
		return fmt.Errorf("failed to open file %s: %w", a.FilePath, err)
	}
	defer file.Close()

	var updatedLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		for pattern, replacement := range a.ReplacePatterns {
			if pattern.MatchString(line) {
				line = pattern.ReplaceAllString(line, replacement)

				// Apply only the first matching replacement
				break
			}
		}

		updatedLines = append(updatedLines, line)
	}
	if err := scanner.Err(); err != nil {
		a.Logger.Error("Failed to read file",
			"FilePath", a.FilePath,
			"error", err,
		)
		return fmt.Errorf("failed to read file %s: %w", a.FilePath, err)
	}

	file, err = os.Create(a.FilePath)
	if err != nil {
		a.Logger.Error("Failed to open file for writing",
			"FilePath", a.FilePath,
			"error", err,
		)
		return fmt.Errorf("failed to open file for writing %s: %w", a.FilePath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range updatedLines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			a.Logger.Error("Failed to write line to file",
				"FilePath", a.FilePath,
				"Line", line,
				"error", err,
			)
			return fmt.Errorf("failed to write line to file %s: %w", a.FilePath, err)
		}
	}
	if err := writer.Flush(); err != nil {
		a.Logger.Error("Failed to flush writer",
			"FilePath", a.FilePath,
			"error", err,
		)
		return fmt.Errorf("failed to flush writer for file %s: %w", a.FilePath, err)
	}

	return nil
}
