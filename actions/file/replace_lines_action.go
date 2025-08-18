package file

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
)

type ReplaceLinesAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder

	FilePath        string
	ReplacePatterns map[*regexp.Regexp]string
	// Optional parameterized replacements; if set, these take precedence over ReplacePatterns
	ReplaceParamPatterns map[*regexp.Regexp]task_engine.ActionParameter
	// Optional file path parameter
	FilePathParam task_engine.ActionParameter
}

// NewReplaceLinesAction creates a new ReplaceLinesAction with the given logger
func NewReplaceLinesAction(logger *slog.Logger) *ReplaceLinesAction {
	return &ReplaceLinesAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
	}
}

// WithParameters sets the parameters for file path and replacement patterns
func (a *ReplaceLinesAction) WithParameters(filePathParam task_engine.ActionParameter, replaceParamPatterns map[*regexp.Regexp]task_engine.ActionParameter) (*task_engine.Action[*ReplaceLinesAction], error) {
	a.FilePathParam = filePathParam
	a.ReplaceParamPatterns = replaceParamPatterns

	return &task_engine.Action[*ReplaceLinesAction]{
		ID:      "replace-lines-action",
		Name:    "Replace Lines",
		Wrapped: a,
	}, nil
}

func (a *ReplaceLinesAction) Execute(ctx context.Context) error {
	// Resolve file path parameter if provided using the ParameterResolver
	if a.FilePathParam != nil {
		pathValue, err := a.ResolveStringParameter(ctx, a.FilePathParam, "file path")
		if err != nil {
			return err
		}
		a.FilePath = pathValue
	}

	if a.FilePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Resolve parameterized replacements first, if provided
	var resolvedReplacements map[*regexp.Regexp]string
	if len(a.ReplaceParamPatterns) > 0 {
		resolvedReplacements = make(map[*regexp.Regexp]string, len(a.ReplaceParamPatterns))
		for pattern, param := range a.ReplaceParamPatterns {
			if param == nil {
				resolvedReplacements[pattern] = ""
				continue
			}
			val, err := a.ResolveParameter(ctx, param, "replacement")
			if err != nil {
				return err
			}
			var replacement string
			switch v := val.(type) {
			case string:
				replacement = v
			case []byte:
				replacement = string(v)
			default:
				replacement = fmt.Sprint(v)
			}
			resolvedReplacements[pattern] = replacement
		}
	} else {
		resolvedReplacements = a.ReplacePatterns
	}

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

		for pattern, replacement := range resolvedReplacements {
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

func (a *ReplaceLinesAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"filePath": a.FilePath,
		"patterns": len(a.ReplacePatterns),
	})
}
