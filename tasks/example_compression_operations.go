package tasks

import (
	"log/slog"

	engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
)

// NewCompressionOperationsTask creates an example task that demonstrates compression and decompression
func NewCompressionOperationsTask(logger *slog.Logger, workingDir string) *engine.Task {
	// Step 1: Create a test file with compressible content
	sourceFile := workingDir + "/test_compression.txt"
	content := "This is a test file with repeated content that should compress well. " +
		"This is a test file with repeated content that should compress well. " +
		"This is a test file with repeated content that should compress well. " +
		"This is a test file with repeated content that should compress well. " +
		"This is a test file with repeated content that should compress well. " +
		"This is a test file with repeated content that should compress well. " +
		"This is a test file with repeated content that should compress well. " +
		"This is a test file with repeated content that should compress well. " +
		"This is a test file with repeated content that should compress well. " +
		"This is a test file with repeated content that should compress well."

	// Step 2: Define compressed and decompressed file paths
	compressedFile := workingDir + "/test_compression.txt.gz"
	decompressedFile := workingDir + "/test_decompressed.txt"

	return &engine.Task{
		ID:   "compression-example",
		Name: "Compression Operations Example",
		Actions: []engine.ActionWrapper{
			// Step 1: Create the test file
			func() engine.ActionWrapper {
				action, err := file.NewWriteFileAction(sourceFile, []byte(content), true, nil, logger)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 2: Compress the file using gzip
			func() engine.ActionWrapper {
				action, err := file.NewCompressFileAction(sourceFile, compressedFile, file.GzipCompression, logger)
				if err != nil {
					logger.Error("Failed to create compress file action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 3: Decompress the file back to a new location
			func() engine.ActionWrapper {
				action, err := file.NewDecompressFileAction(compressedFile, decompressedFile, file.GzipCompression, logger)
				if err != nil {
					logger.Error("Failed to create decompress file action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// NewCompressionWithAutoDetectTask creates an example task that demonstrates auto-detection
func NewCompressionWithAutoDetectTask(logger *slog.Logger, workingDir string) *engine.Task {
	// Create a test file
	sourceFile := workingDir + "/auto_detect_test.txt"
	content := "Test content for auto-detection example"

	// Compress with .gz extension
	compressedFile := workingDir + "/auto_detect_test.txt.gz"

	// Decompress using auto-detection (empty compression type)
	decompressedFile := workingDir + "/auto_detect_decompressed.txt"

	return &engine.Task{
		ID:   "compression-auto-detect",
		Name: "Compression Auto-Detection Example",
		Actions: []engine.ActionWrapper{
			// Create the test file
			func() engine.ActionWrapper {
				action, err := file.NewWriteFileAction(sourceFile, []byte(content), true, nil, logger)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),

			// Compress with .gz extension
			func() engine.ActionWrapper {
				action, err := file.NewCompressFileAction(sourceFile, compressedFile, file.GzipCompression, logger)
				if err != nil {
					logger.Error("Failed to create compress file action", "error", err)
					return nil
				}
				return action
			}(),

			// Decompress using auto-detection (empty compression type)
			func() engine.ActionWrapper {
				action, err := file.NewDecompressFileAction(compressedFile, decompressedFile, "", logger)
				if err != nil {
					logger.Error("Failed to create decompress file action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// NewCompressionWorkflowTask creates an example task that demonstrates a complex compression workflow
func NewCompressionWorkflowTask(logger *slog.Logger, workingDir string) *engine.Task {
	// Define source files and their contents
	sourceFiles := []string{
		workingDir + "/source1.txt",
		workingDir + "/source2.txt",
		workingDir + "/source3.txt",
	}

	contents := []string{
		"Content for file 1 with repeated patterns for good compression",
		"Content for file 2 with repeated patterns for good compression",
		"Content for file 3 with repeated patterns for good compression",
	}

	// Define compressed file paths
	compressedFiles := []string{
		workingDir + "/source1.txt.gz",
		workingDir + "/source2.txt.gz",
		workingDir + "/source3.txt.gz",
	}

	// Define backup directory
	backupDir := workingDir + "/backup"

	// Create actions slice
	var actions []engine.ActionWrapper

	// Step 1: Create source files
	for i, sourceFile := range sourceFiles {
		content := contents[i] + " " + contents[i] + " " + contents[i] // Repeat for better compression
		action, err := file.NewWriteFileAction(sourceFile, []byte(content), true, nil, logger)
		if err != nil {
			logger.Error("Failed to create write file action", "error", err)
			continue
		}
		actions = append(actions, action)
	}

	// Step 2: Compress each file
	for i, sourceFile := range sourceFiles {
		action, err := file.NewCompressFileAction(sourceFile, compressedFiles[i], file.GzipCompression, logger)
		if err != nil {
			logger.Error("Failed to create compress file action", "error", err)
			continue
		}
		actions = append(actions, action)
	}

	// Step 3: Decompress all files to a backup directory
	for i, compressedFile := range compressedFiles {
		decompressedFile := backupDir + "/restored" + string(rune('1'+i)) + ".txt"
		action, err := file.NewDecompressFileAction(compressedFile, decompressedFile, "", logger)
		if err != nil {
			logger.Error("Failed to create decompress file action", "error", err)
			continue
		}
		actions = append(actions, action)
	}

	return &engine.Task{
		ID:      "compression-workflow",
		Name:    "Compression Workflow Example",
		Actions: actions,
		Logger:  logger,
	}
}
