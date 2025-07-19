package file

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	engine "github.com/ndizazzo/task-engine"
	task_engine "github.com/ndizazzo/task-engine"
)

// NewDecompressFileAction creates an action that decompresses a file using the specified compression type.
// If compressionType is empty, it will be auto-detected from the file extension.
func NewDecompressFileAction(sourcePath string, destinationPath string, compressionType CompressionType, logger *slog.Logger) *engine.Action[*DecompressFileAction] {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	if sourcePath == "" {
		logger.Error("Invalid parameter: sourcePath cannot be empty")
		return nil
	}
	if destinationPath == "" {
		logger.Error("Invalid parameter: destinationPath cannot be empty")
		return nil
	}

	// Auto-detect compression type if not specified
	if compressionType == "" {
		compressionType = DetectCompressionType(sourcePath)
		if compressionType == "" {
			logger.Error("Could not auto-detect compression type from file extension", "sourcePath", sourcePath)
			return nil
		}
	}

	// Validate compression type
	switch compressionType {
	case GzipCompression:
		// Valid compression type
	default:
		logger.Error("Invalid compression type", "compressionType", compressionType)
		return nil
	}

	id := fmt.Sprintf("decompress-file-%s-%s", compressionType, filepath.Base(sourcePath))
	return &task_engine.Action[*DecompressFileAction]{
		ID: id,
		Wrapped: &DecompressFileAction{
			BaseAction:      task_engine.BaseAction{Logger: logger},
			SourcePath:      sourcePath,
			DestinationPath: destinationPath,
			CompressionType: compressionType,
		},
	}
}

// DecompressFileAction decompresses a file using the specified compression algorithm
type DecompressFileAction struct {
	engine.BaseAction
	SourcePath      string
	DestinationPath string
	CompressionType CompressionType
}

func (a *DecompressFileAction) Execute(execCtx context.Context) error {
	a.Logger.Info("Attempting to decompress file",
		"source", a.SourcePath,
		"destination", a.DestinationPath,
		"compressionType", a.CompressionType)

	// Check if source file exists
	sourceInfo, err := os.Stat(a.SourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			errMsg := fmt.Sprintf("source file %s does not exist", a.SourcePath)
			a.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
		a.Logger.Error("Failed to stat source file", "path", a.SourcePath, "error", err)
		return fmt.Errorf("failed to stat source file %s: %w", a.SourcePath, err)
	}

	// Check if it's a regular file
	if sourceInfo.IsDir() {
		errMsg := fmt.Sprintf("source path %s is a directory, not a file", a.SourcePath)
		a.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	// Create destination directory if needed
	destDir := filepath.Dir(a.DestinationPath)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		a.Logger.Error("Failed to create destination directory", "path", destDir, "error", err)
		return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}

	// Open source file
	sourceFile, err := os.Open(a.SourcePath)
	if err != nil {
		a.Logger.Error("Failed to open source file", "path", a.SourcePath, "error", err)
		return fmt.Errorf("failed to open source file %s: %w", a.SourcePath, err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(a.DestinationPath)
	if err != nil {
		a.Logger.Error("Failed to create destination file", "path", a.DestinationPath, "error", err)
		return fmt.Errorf("failed to create destination file %s: %w", a.DestinationPath, err)
	}
	defer destFile.Close()

	// Decompress based on compression type
	switch a.CompressionType {
	case GzipCompression:
		err = a.decompressGzip(sourceFile, destFile)
	default:
		err = fmt.Errorf("unsupported compression type: %s", a.CompressionType)
	}

	if err != nil {
		a.Logger.Error("Failed to decompress file", "source", a.SourcePath, "destination", a.DestinationPath, "error", err)
		return fmt.Errorf("failed to decompress file %s to %s: %w", a.SourcePath, a.DestinationPath, err)
	}

	// Get decompressed file size
	destInfo, err := os.Stat(a.DestinationPath)
	if err != nil {
		a.Logger.Warn("Failed to get decompressed file size", "path", a.DestinationPath, "error", err)
	} else {
		compressionRatio := float64(sourceInfo.Size()) / float64(destInfo.Size()) * 100
		a.Logger.Info("Successfully decompressed file",
			"source", a.SourcePath,
			"destination", a.DestinationPath,
			"compressedSize", sourceInfo.Size(),
			"decompressedSize", destInfo.Size(),
			"compressionRatio", fmt.Sprintf("%.1f%%", compressionRatio))
	}

	return nil
}

// decompressGzip decompresses a file using gzip decompression
func (a *DecompressFileAction) decompressGzip(source io.Reader, destination io.Writer) error {
	gzipReader, err := gzip.NewReader(source)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	_, err = io.Copy(destination, gzipReader)
	if err != nil {
		return fmt.Errorf("failed to decompress with gzip: %w", err)
	}

	return nil
}

// DetectCompressionType auto-detects the compression type from file extension
func DetectCompressionType(filePath string) CompressionType {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".gz", ".gzip":
		return GzipCompression
	// Future compression types can be added here:
	// case ".z", ".zz":
	//     return ZlibCompression
	// case ".lz4":
	//     return Lz4Compression
	default:
		return ""
	}
}
