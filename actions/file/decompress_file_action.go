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
)

// NewDecompressFileAction creates a new DecompressFileAction with the given logger
func NewDecompressFileAction(logger *slog.Logger) *DecompressFileAction {
	return &DecompressFileAction{
		BaseAction: engine.NewBaseAction(logger),
	}
}

// WithParameters sets the parameters for source path, destination path, and compression type
func (a *DecompressFileAction) WithParameters(sourcePathParam, destinationPathParam engine.ActionParameter, compressionType CompressionType) (*engine.Action[*DecompressFileAction], error) {
	// Validate compression type if specified
	if compressionType != "" {
		switch compressionType {
		case GzipCompression:
			// Valid compression type
		default:
			return nil, fmt.Errorf("invalid compression type: %s", compressionType)
		}
	}

	a.SourcePathParam = sourcePathParam
	a.DestinationPathParam = destinationPathParam
	a.CompressionType = compressionType

	return &engine.Action[*DecompressFileAction]{
		ID:      "decompress-file-action",
		Name:    "Decompress File",
		Wrapped: a,
	}, nil
}

// DecompressFileAction decompresses a file using the specified compression algorithm
type DecompressFileAction struct {
	engine.BaseAction
	SourcePath      string
	DestinationPath string
	CompressionType CompressionType

	// Parameter-aware fields
	SourcePathParam      engine.ActionParameter
	DestinationPathParam engine.ActionParameter
}

func (a *DecompressFileAction) Execute(execCtx context.Context) error {
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

	// Auto-detect compression type if not specified
	if a.CompressionType == "" {
		a.CompressionType = DetectCompressionType(a.SourcePath)
		if a.CompressionType == "" {
			return fmt.Errorf("could not auto-detect compression type from file extension: %s", a.SourcePath)
		}
	}

	a.Logger.Info("Attempting to decompress file",
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

	// Use a limited reader to prevent decompression bomb attacks
	// Limit to 100MB to prevent DoS attacks
	const maxDecompressedSize = 100 * 1024 * 1024 // 100MB
	limitedReader := io.LimitReader(gzipReader, maxDecompressedSize)

	_, err = io.Copy(destination, limitedReader)
	if err != nil {
		return fmt.Errorf("failed to decompress with gzip: %w", err)
	}

	return nil
}

// GetOutput returns metadata about the decompression operation
func (a *DecompressFileAction) GetOutput() interface{} {
	return map[string]interface{}{
		"source":          a.SourcePath,
		"destination":     a.DestinationPath,
		"compressionType": string(a.CompressionType),
		"success":         true,
	}
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
