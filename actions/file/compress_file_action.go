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

// NewCompressFileAction creates a new CompressFileAction with the given logger
func NewCompressFileAction(logger *slog.Logger) *CompressFileAction {
	return &CompressFileAction{
		BaseAction: engine.NewBaseAction(logger),
	}
}

// WithParameters sets the parameters for source path, destination path, and compression type
func (a *CompressFileAction) WithParameters(sourcePathParam, destinationPathParam engine.ActionParameter, compressionType CompressionType) (*engine.Action[*CompressFileAction], error) {
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

	a.SourcePathParam = sourcePathParam
	a.DestinationPathParam = destinationPathParam
	a.CompressionType = compressionType

	return &engine.Action[*CompressFileAction]{
		ID:      "compress-file-action",
		Name:    "Compress File",
		Wrapped: a,
	}, nil
}

// CompressFileAction compresses a file using the specified compression algorithm
type CompressFileAction struct {
	engine.BaseAction
	SourcePath      string
	DestinationPath string
	CompressionType CompressionType

	// Parameter-aware fields
	SourcePathParam      engine.ActionParameter
	DestinationPathParam engine.ActionParameter
}

func (a *CompressFileAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *engine.GlobalContext
	if gc, ok := execCtx.Value(engine.GlobalContextKey).(*engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve parameters if they exist
	if a.SourcePathParam != nil {
		sourceValue, err := a.SourcePathParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve source path parameter: %w", err)
		}
		if sourceStr, ok := sourceValue.(string); ok {
			a.SourcePath = sourceStr
		} else {
			return fmt.Errorf("source path parameter is not a string, got %T", sourceValue)
		}
	}

	if a.DestinationPathParam != nil {
		destValue, err := a.DestinationPathParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve destination path parameter: %w", err)
		}
		if destStr, ok := destValue.(string); ok {
			a.DestinationPath = destStr
		} else {
			return fmt.Errorf("destination path parameter is not a string, got %T", destValue)
		}
	}

	a.Logger.Info("Attempting to compress file",
		"source", a.SourcePath,
		"destination", a.DestinationPath,
		"compressionType", a.CompressionType)
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
	if sourceInfo.IsDir() {
		errMsg := fmt.Sprintf("source path %s is a directory, not a file", a.SourcePath)
		a.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	// Create destination directory if needed
	destDir := filepath.Dir(a.DestinationPath)
	if err := os.MkdirAll(destDir, 0o750); err != nil {
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

// GetOutput returns metadata about the compression operation
func (a *CompressFileAction) GetOutput() interface{} {
	return map[string]interface{}{
		"source":          a.SourcePath,
		"destination":     a.DestinationPath,
		"compressionType": string(a.CompressionType),
		"success":         true,
	}
}
