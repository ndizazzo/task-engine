package tasks

import (
	"log/slog"
	"regexp"

	engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
)

// NewFileOperationsTask creates an example task that demonstrates
// various file operations including creating, copying, replacing text, and cleanup
func NewFileOperationsTask(logger *slog.Logger, workingDir string) *engine.Task {
	return &engine.Task{
		ID:   "file-operations-example",
		Name: "File Operations Example",
		Actions: []engine.ActionWrapper{
			// Step 1: Create project structure
			func() engine.ActionWrapper {
				action, err := file.NewCreateDirectoriesAction(
					logger,
					workingDir,
					[]string{"src", "tests", "docs", "tmp"},
				)
				if err != nil {
					logger.Error("Failed to create directories action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 2: Create initial source file
			func() engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					workingDir+"/src/main.go",
					[]byte(initialSourceCode),
					true,
					nil,
					logger,
				)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 3: Create a configuration file
			func() engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					workingDir+"/config.json",
					[]byte(initialConfig),
					true,
					nil,
					logger,
				)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 4: Copy the source file to backup
			func() engine.ActionWrapper {
				action, err := file.NewCopyFileAction(
					workingDir+"/src/main.go",
					workingDir+"/src/main.go.backup",
					true,  // createDir
					false, // recursive
					logger,
				)
				if err != nil {
					logger.Error("Failed to create copy file action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 5: Replace placeholder text in the source file
			file.NewReplaceLinesAction(
				workingDir+"/src/main.go",
				map[*regexp.Regexp]string{
					regexp.MustCompile("TODO: implement main logic"): "fmt.Println(\"Hello, Task Engine!\")",
				},
				logger,
			),

			// Step 6: Replace configuration values
			file.NewReplaceLinesAction(
				workingDir+"/config.json",
				map[*regexp.Regexp]string{
					regexp.MustCompile("\"development\""): "\"production\"",
				},
				logger,
			),

			// Step 7: Create documentation
			func() engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					workingDir+"/docs/README.md",
					[]byte(documentationContent),
					true,
					nil,
					logger,
				)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 8: Create a temporary test file
			func() engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					workingDir+"/tmp/test.txt",
					[]byte("This is a temporary test file"),
					true,
					nil,
					logger,
				)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 9: Clean up temporary file
			func() engine.ActionWrapper {
				action, err := file.NewDeletePathAction(
					workingDir+"/tmp/test.txt",
					false, // recursive
					false, // dryRun
					logger,
				)
				if err != nil {
					logger.Error("Failed to create delete path action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

const initialSourceCode = `package main

import "fmt"

func main() {
	// TODO: implement main logic
	fmt.Println("Starting application...")
}
`

const initialConfig = `{
	"app_name": "example-app",
	"version": "1.0.0",
	"environment": "development",
	"database": {
		"host": "localhost",
		"port": 5432,
		"name": "example_db"
	},
	"logging": {
		"level": "info",
		"format": "json"
	}
}
`

const documentationContent = `# Example Application

This is an example application demonstrating the Task Engine capabilities.

## Features

- File operations and management
- Configuration handling
- Directory structure creation
- Text replacement and processing

## Usage

Run the application with:

` + "```bash" + `
go run main.go
` + "```" + `

## Configuration

The application uses a JSON configuration file located at ` + "`config.json`" + `.

## Development

This project demonstrates:

1. **Directory Structure**: Organized project layout
2. **File Operations**: Create, copy, modify, and delete files
3. **Text Processing**: Replace content in files
4. **Cleanup**: Remove temporary files

## License

This is an example project for demonstration purposes.
`
