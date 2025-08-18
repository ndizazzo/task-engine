package file

import (
	"archive/tar"
	"archive/zip"
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

// ArchiveType represents the type of archive to extract
type ArchiveType string

const (
	// TarArchive represents tar archive
	TarArchive ArchiveType = "tar"
	// TarGzArchive represents tar.gz archive (requires decompression first)
	TarGzArchive ArchiveType = "tar.gz"
	// ZipArchive represents zip archive
	ZipArchive ArchiveType = "zip"
)

// NewExtractFileAction creates a new ExtractFileAction with the given logger
func NewExtractFileAction(logger *slog.Logger) *ExtractFileAction {
	return &ExtractFileAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
	}
}

// WithParameters sets the parameters for source and destination paths and archive type
func (a *ExtractFileAction) WithParameters(
	sourcePathParam task_engine.ActionParameter,
	destinationPathParam task_engine.ActionParameter,
	archiveType ArchiveType,
) (*task_engine.Action[*ExtractFileAction], error) {
	// Validate archive type if specified
	if archiveType != "" {
		switch archiveType {
		case TarArchive, ZipArchive, TarGzArchive:
			// Valid archive type
		default:
			return nil, fmt.Errorf("invalid archive type: %s", archiveType)
		}
	}

	a.SourcePathParam = sourcePathParam
	a.DestinationPathParam = destinationPathParam
	a.ArchiveType = archiveType

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*ExtractFileAction](a.Logger)
	return constructor.WrapAction(a, "Extract File", "extract-file-action"), nil
}

// ExtractFileAction extracts an archive to the specified destination
type ExtractFileAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	SourcePath      string
	DestinationPath string
	ArchiveType     ArchiveType

	// Parameter-aware fields
	SourcePathParam      task_engine.ActionParameter
	DestinationPathParam task_engine.ActionParameter
}

func (a *ExtractFileAction) Execute(execCtx context.Context) error {
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

	if a.SourcePath == "" {
		return fmt.Errorf("source path cannot be empty")
	}

	if a.DestinationPath == "" {
		return fmt.Errorf("destination path cannot be empty")
	}

	// Auto-detect archive type if not specified
	if a.ArchiveType == "" {
		a.ArchiveType = DetectArchiveType(a.SourcePath)
		if a.ArchiveType == "" {
			return fmt.Errorf("could not auto-detect archive type from file extension: %s", a.SourcePath)
		}
	}

	a.Logger.Info("Attempting to extract archive",
		"source", a.SourcePath,
		"destination", a.DestinationPath,
		"archiveType", a.ArchiveType)
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
	if err := os.MkdirAll(a.DestinationPath, 0o750); err != nil {
		a.Logger.Error("Failed to create destination directory", "path", a.DestinationPath, "error", err)
		return fmt.Errorf("failed to create destination directory %s: %w", a.DestinationPath, err)
	}
	if a.ArchiveType == TarGzArchive {
		if isCompressed, compressionType := a.detectCompression(a.SourcePath); isCompressed {
			errMsg := fmt.Sprintf("file %s is compressed with %s. Please decompress it first using DecompressFileAction, then extract using ExtractFileAction", a.SourcePath, compressionType)
			a.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
	}

	// Open source file
	sourceFile, err := os.Open(a.SourcePath)
	if err != nil {
		a.Logger.Error("Failed to open source file", "path", a.SourcePath, "error", err)
		return fmt.Errorf("failed to open source file %s: %w", a.SourcePath, err)
	}
	defer sourceFile.Close()

	// Extract based on archive type
	switch a.ArchiveType {
	case TarArchive, TarGzArchive:
		err = a.extractTar(sourceFile, a.DestinationPath)
	case ZipArchive:
		err = a.extractZip(sourceFile, a.DestinationPath)
	default:
		err = fmt.Errorf("unsupported archive type: %s", a.ArchiveType)
	}

	if err != nil {
		a.Logger.Error("Failed to extract archive", "source", a.SourcePath, "destination", a.DestinationPath, "error", err)
		return fmt.Errorf("failed to extract archive %s to %s: %w", a.SourcePath, a.DestinationPath, err)
	}

	a.Logger.Info("Successfully extracted archive",
		"source", a.SourcePath,
		"destination", a.DestinationPath,
		"archiveType", a.ArchiveType)

	return nil
}

// validateAndSanitizePath validates and sanitizes a file path to prevent path traversal attacks
func (a *ExtractFileAction) validateAndSanitizePath(fileName, destination string) (string, error) {
	// Sanitize the file name to prevent path traversal
	sanitizedName := filepath.Clean(fileName)
	if strings.Contains(sanitizedName, "..") {
		return "", fmt.Errorf("illegal file path: %s", fileName)
	}

	targetPath := filepath.Join(destination, sanitizedName)
	if !strings.HasPrefix(targetPath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return "", fmt.Errorf("illegal file path: %s", fileName)
	}

	return targetPath, nil
}

// createTargetFile creates a target file and ensures its directory exists
func (a *ExtractFileAction) createTargetFile(targetPath string) (*os.File, error) {
	// Ensure the target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", targetPath, err)
	}

	return targetFile, nil
}

// copyWithLimit copies data from reader to file with a size limit to prevent decompression bombs
func (a *ExtractFileAction) copyWithLimit(dst *os.File, src io.Reader, fileName string) error {
	limitedReader := io.LimitReader(src, 100*1024*1024) // 100MB limit
	if _, err := io.Copy(dst, limitedReader); err != nil {
		return fmt.Errorf("failed to copy file content for %s: %w", fileName, err)
	}
	return nil
}

// setFilePermissions safely sets file permissions with overflow protection
func (a *ExtractFileAction) setFilePermissions(targetPath string, mode int64) {
	// Use safe conversion to avoid integer overflow
	safeMode := mode & 0o777                          // Only use the permission bits, avoid overflow
	fileMode := os.FileMode(uint32(safeMode & 0x1FF)) // Ensure only 9 bits are used
	if err := os.Chmod(targetPath, fileMode); err != nil {
		a.Logger.Warn("Failed to set file permissions", "file", targetPath, "error", err)
	}
}

// extractTar extracts a tar archive
func (a *ExtractFileAction) extractTar(source io.Reader, destination string) error {
	tarReader := tar.NewReader(source)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip if it's a directory (tar creates directories automatically)
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Validate and sanitize path
		targetPath, err := a.validateAndSanitizePath(header.Name, destination)
		if err != nil {
			return err
		}
		targetFile, err := a.createTargetFile(targetPath)
		if err != nil {
			return err
		}

		// Copy file content with size limit
		if err := a.copyWithLimit(targetFile, tarReader, header.Name); err != nil {
			_ = targetFile.Close()
			return err
		}

		_ = targetFile.Close()

		// Set file permissions
		a.setFilePermissions(targetPath, header.Mode)
	}

	return nil
}

// extractZip extracts a zip archive
func (a *ExtractFileAction) extractZip(source io.Reader, destination string) error {
	// For zip files, we need to read the entire file into memory first
	// since zip.Reader requires a io.ReaderAt
	data, err := io.ReadAll(source)
	if err != nil {
		return fmt.Errorf("failed to read zip file: %w", err)
	}

	// Create a zip reader
	zipReader, err := zip.NewReader(strings.NewReader(string(data)), int64(len(data)))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	// Extract each file in the zip
	for _, file := range zipReader.File {
		// Validate and sanitize path
		targetPath, err := a.validateAndSanitizePath(file.Name, destination)
		if err != nil {
			return err
		}

		// If it's a directory, create it and continue
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o750); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
			continue
		}
		targetFile, err := a.createTargetFile(targetPath)
		if err != nil {
			return err
		}

		// Open the zip file
		zipFile, err := file.Open()
		if err != nil {
			_ = targetFile.Close()
			return fmt.Errorf("failed to open zip file %s: %w", file.Name, err)
		}

		// Copy file content with size limit
		if err := a.copyWithLimit(targetFile, zipFile, file.Name); err != nil {
			_ = zipFile.Close()
			_ = targetFile.Close()
			return err
		}

		_ = zipFile.Close()
		_ = targetFile.Close()

		// Set file permissions
		a.setFilePermissions(targetPath, int64(file.Mode()))
	}

	return nil
}

// GetOutput returns metadata about the extraction operation
func (a *ExtractFileAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"source":      a.SourcePath,
		"destination": a.DestinationPath,
		"archiveType": string(a.ArchiveType),
	})
}

// detectCompression checks if a file is compressed and returns the compression type
func (a *ExtractFileAction) detectCompression(filePath string) (bool, string) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, ""
	}
	defer file.Close()

	// Try to create a gzip reader to test if it's gzip compressed
	_, err = gzip.NewReader(file)
	if err == nil {
		return true, "gzip"
	}

	// Reset file position for other checks
	_, _ = file.Seek(0, 0)

	// Read first few bytes to check for gzip magic number
	buffer := make([]byte, 2)
	_, err = file.Read(buffer)
	if err != nil {
		return false, ""
	}
	if buffer[0] == 0x1f && buffer[1] == 0x8b {
		return true, "gzip"
	}

	return false, ""
}

// DetectArchiveType auto-detects the archive type from file extension
func DetectArchiveType(filePath string) ArchiveType {
	ext := strings.ToLower(filepath.Ext(filePath))
	baseName := strings.ToLower(filepath.Base(filePath))

	switch {
	case ext == ".tar":
		return TarArchive
	case ext == ".gz" && strings.HasSuffix(baseName, ".tar.gz"):
		return TarGzArchive
	case ext == ".zip":
		return ZipArchive
	default:
		return ""
	}
}
