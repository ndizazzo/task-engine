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
			file.NewWriteFileAction(sourceFile, []byte(content), true, nil, logger),

			// Step 2: Compress the file using gzip
			file.NewCompressFileAction(sourceFile, compressedFile, file.GzipCompression, logger),

			// Step 3: Decompress the file back to a new location
			file.NewDecompressFileAction(compressedFile, decompressedFile, file.GzipCompression, logger),
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
			file.NewWriteFileAction(sourceFile, []byte(content), true, nil, logger),

			// Compress with .gz extension
			file.NewCompressFileAction(sourceFile, compressedFile, file.GzipCompression, logger),

			// Decompress using auto-detection (empty compression type)
			file.NewDecompressFileAction(compressedFile, decompressedFile, "", logger),
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
		actions = append(actions, file.NewWriteFileAction(sourceFile, []byte(content), true, nil, logger))
	}

	// Step 2: Compress each file
	for i, sourceFile := range sourceFiles {
		actions = append(actions, file.NewCompressFileAction(sourceFile, compressedFiles[i], file.GzipCompression, logger))
	}

	// Step 3: Decompress all files to a backup directory
	for i, compressedFile := range compressedFiles {
		decompressedFile := backupDir + "/restored" + string(rune('1'+i)) + ".txt"
		actions = append(actions, file.NewDecompressFileAction(compressedFile, decompressedFile, "", logger))
	}

	return &engine.Task{
		ID:      "compression-workflow",
		Name:    "Compression Workflow Example",
		Actions: actions,
		Logger:  logger,
	}
}
