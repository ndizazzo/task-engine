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

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
)

// NewDecompressFileAction creates a new DecompressFileAction with the given logger
func NewDecompressFileAction(logger *slog.Logger) *DecompressFileAction {
	return &DecompressFileAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
	}
}

// WithParameters sets the parameters for source path, destination path, and compression type
func (a *DecompressFileAction) WithParameters(
	sourcePathParam task_engine.ActionParameter,
	destinationPathParam task_engine.ActionParameter,
	compressionType CompressionType,
) (*task_engine.Action[*DecompressFileAction], error) {
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

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*DecompressFileAction](a.Logger)
	return constructor.WrapAction(a, "Decompress File", "decompress-file-action"), nil
}

// DecompressFileAction decompresses a file using the specified compression algorithm
type DecompressFileAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	SourcePath      string
	DestinationPath string
	CompressionType CompressionType

	// Parameter-aware fields
	SourcePathParam      task_engine.ActionParameter
	DestinationPathParam task_engine.ActionParameter
}

func (a *DecompressFileAction) Execute(execCtx context.Context) error {
	// Resolve parameters using the ParameterResolver
	if a.SourcePathParam != nil {
		sourceValue, err := a.ResolveStringParameter(execCtx, a.SourcePathParam, "source path")
		if err != nil {
			return err
		}
		a.SourcePath = sourceValue
	}

	if a.DestinationPathParam != nil {
		destValue, err := a.ResolveStringParameter(execCtx, a.DestinationPathParam, "destination path")
		if err != nil {
			return err
		}
		a.DestinationPath = destValue
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
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"source":          a.SourcePath,
		"destination":     a.DestinationPath,
		"compressionType": string(a.CompressionType),
	})
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
