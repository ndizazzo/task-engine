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

	engine "github.com/ndizazzo/task-engine"
)

// CompressionType represents the type of compression to use
type CompressionType string

const (
	// GzipCompression represents gzip compression
	GzipCompression CompressionType = "gzip"
	// Future compression types can be added here:
	// ZlibCompression CompressionType = "zlib"
	// Lz4Compression  CompressionType = "lz4"
)

// NewCompressFileAction creates an action that compresses a file using the specified compression type.
func NewCompressFileAction(sourcePath string, destinationPath string, compressionType CompressionType, logger *slog.Logger) (*engine.Action[*CompressFileAction], error) {
	if sourcePath == "" {
		return nil, fmt.Errorf("invalid parameter: sourcePath cannot be empty")
	}
	if destinationPath == "" {
		return nil, fmt.Errorf("invalid parameter: destinationPath cannot be empty")
	}
	if compressionType == "" {
		return nil, fmt.Errorf("invalid parameter: compressionType cannot be empty")
	}

	// Validate compression type
	switch compressionType {
	case GzipCompression:
		// Valid compression type
	default:
		return nil, fmt.Errorf("invalid compression type: %s", compressionType)
	}

	id := fmt.Sprintf("compress-file-%s-%s", compressionType, filepath.Base(sourcePath))
	return &engine.Action[*CompressFileAction]{
		ID: id,
		Wrapped: &CompressFileAction{
			BaseAction:      engine.BaseAction{Logger: logger},
			SourcePath:      sourcePath,
			DestinationPath: destinationPath,
			CompressionType: compressionType,
		},
	}, nil
}

// CompressFileAction compresses a file using the specified compression algorithm
type CompressFileAction struct {
	engine.BaseAction
	SourcePath      string
	DestinationPath string
	CompressionType CompressionType
}

func (a *CompressFileAction) Execute(execCtx context.Context) error {
	a.Logger.Info("Attempting to compress file",
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

	// Compress based on compression type
	switch a.CompressionType {
	case GzipCompression:
		err = a.compressGzip(sourceFile, destFile)
	default:
		err = fmt.Errorf("unsupported compression type: %s", a.CompressionType)
	}

	if err != nil {
		a.Logger.Error("Failed to compress file", "source", a.SourcePath, "destination", a.DestinationPath, "error", err)
		return fmt.Errorf("failed to compress file %s to %s: %w", a.SourcePath, a.DestinationPath, err)
	}

	// Get compressed file size
	destInfo, err := os.Stat(a.DestinationPath)
	if err != nil {
		a.Logger.Warn("Failed to get compressed file size", "path", a.DestinationPath, "error", err)
	} else {
		compressionRatio := float64(destInfo.Size()) / float64(sourceInfo.Size()) * 100
		a.Logger.Info("Successfully compressed file",
			"source", a.SourcePath,
			"destination", a.DestinationPath,
			"originalSize", sourceInfo.Size(),
			"compressedSize", destInfo.Size(),
			"compressionRatio", fmt.Sprintf("%.1f%%", compressionRatio))
	}

	return nil
}

// compressGzip compresses a file using gzip compression
func (a *CompressFileAction) compressGzip(source io.Reader, destination io.Writer) error {
	gzipWriter := gzip.NewWriter(destination)
	defer gzipWriter.Close()

	_, err := io.Copy(gzipWriter, source)
	if err != nil {
		return fmt.Errorf("failed to compress with gzip: %w", err)
	}

	return nil
}
