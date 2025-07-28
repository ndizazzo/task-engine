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

	engine "github.com/ndizazzo/task-engine"
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

// NewExtractFileAction creates an action that extracts an archive to the specified destination.
// If archiveType is empty, it will be auto-detected from the file extension.
// Note: For compressed archives like .tar.gz, use DecompressFileAction first, then ExtractFileAction.
func NewExtractFileAction(sourcePath string, destinationPath string, archiveType ArchiveType, logger *slog.Logger) (*engine.Action[*ExtractFileAction], error) {
	if sourcePath == "" {
		return nil, fmt.Errorf("invalid parameter: sourcePath cannot be empty")
	}
	if destinationPath == "" {
		return nil, fmt.Errorf("invalid parameter: destinationPath cannot be empty")
	}

	// Auto-detect archive type if not specified
	if archiveType == "" {
		archiveType = DetectArchiveType(sourcePath)
		if archiveType == "" {
			return nil, fmt.Errorf("could not auto-detect archive type from file extension: %s", sourcePath)
		}
	}

	// Validate archive type
	switch archiveType {
	case TarArchive, TarGzArchive, ZipArchive:
		// Valid archive type
	default:
		return nil, fmt.Errorf("invalid archive type: %s", archiveType)
	}

	id := fmt.Sprintf("extract-file-%s-%s", archiveType, filepath.Base(sourcePath))
	return &engine.Action[*ExtractFileAction]{
		ID: id,
		Wrapped: &ExtractFileAction{
			BaseAction:      engine.BaseAction{Logger: logger},
			SourcePath:      sourcePath,
			DestinationPath: destinationPath,
			ArchiveType:     archiveType,
		},
	}, nil
}

// ExtractFileAction extracts an archive to the specified destination
type ExtractFileAction struct {
	engine.BaseAction
	SourcePath      string
	DestinationPath string
	ArchiveType     ArchiveType
}

func (a *ExtractFileAction) Execute(execCtx context.Context) error {
	a.Logger.Info("Attempting to extract archive",
		"source", a.SourcePath,
		"destination", a.DestinationPath,
		"archiveType", a.ArchiveType)

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
	if err := os.MkdirAll(a.DestinationPath, 0750); err != nil {
		a.Logger.Error("Failed to create destination directory", "path", a.DestinationPath, "error", err)
		return fmt.Errorf("failed to create destination directory %s: %w", a.DestinationPath, err)
	}

	// Check if the file is compressed and provide helpful error
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
	case TarArchive:
		err = a.extractTar(sourceFile, a.DestinationPath)
	case TarGzArchive:
		err = a.extractTarGz(sourceFile, a.DestinationPath)
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

		// Create the full path for the file
		targetPath := filepath.Join(destination, header.Name)

		// Ensure the target directory exists
		targetDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(targetDir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		// Create the file
		targetFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}

		// Copy the file content
		if _, err := io.Copy(targetFile, tarReader); err != nil {
			targetFile.Close()
			return fmt.Errorf("failed to copy file content for %s: %w", header.Name, err)
		}

		targetFile.Close()

		// Set file permissions
		if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
			a.Logger.Warn("Failed to set file permissions", "file", targetPath, "error", err)
		}
	}

	return nil
}

// extractTarGz extracts a tar.gz archive
// Note: This method expects the file to already be decompressed.
// For compressed .tar.gz files, use DecompressFileAction first, then ExtractFileAction.
func (a *ExtractFileAction) extractTarGz(source io.Reader, destination string) error {
	// For .tar.gz files, we expect the file to already be decompressed
	// The source should be a tar stream, not a compressed stream
	return a.extractTar(source, destination)
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
		// Create the full path for the file
		targetPath := filepath.Join(destination, file.Name)

		// Check for zip slip vulnerability
		if !strings.HasPrefix(targetPath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		// If it's a directory, create it and continue
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
			continue
		}

		// Ensure the target directory exists
		targetDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(targetDir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		// Create the file
		targetFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}

		// Open the zip file
		zipFile, err := file.Open()
		if err != nil {
			targetFile.Close()
			return fmt.Errorf("failed to open zip file %s: %w", file.Name, err)
		}

		// Copy the file content
		if _, err := io.Copy(targetFile, zipFile); err != nil {
			zipFile.Close()
			targetFile.Close()
			return fmt.Errorf("failed to copy file content for %s: %w", file.Name, err)
		}

		zipFile.Close()
		targetFile.Close()

		// Set file permissions
		if err := os.Chmod(targetPath, file.Mode()); err != nil {
			a.Logger.Warn("Failed to set file permissions", "file", targetPath, "error", err)
		}
	}

	return nil
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
	file.Seek(0, 0)

	// Read first few bytes to check for gzip magic number
	buffer := make([]byte, 2)
	_, err = file.Read(buffer)
	if err != nil {
		return false, ""
	}

	// Check for gzip magic number (0x1f 0x8b)
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
