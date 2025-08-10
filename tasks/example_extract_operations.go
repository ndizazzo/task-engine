package tasks

import (
	"archive/tar"
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	engine "github.com/ndizazzo/task-engine"
	fileActions "github.com/ndizazzo/task-engine/actions/file"
)

// NewExtractOperationsTask demonstrates basic archive extraction operations
func NewExtractOperationsTask(logger *slog.Logger) *engine.Task {
	return &engine.Task{
		Name: "extract-operations",
		Actions: []engine.ActionWrapper{
			// Create a simple tar archive
			&engine.Action[*CreateArchiveAction]{
				ID: "create-tar-archive",
				Wrapped: &CreateArchiveAction{
					BaseAction:  engine.BaseAction{Logger: logger},
					SourceDir:   "testdata",
					DestPath:    "testdata.tar",
					ArchiveType: fileActions.TarArchive,
				},
			},
			// Extract the tar archive
			&engine.Action[*fileActions.ExtractFileAction]{
				ID: "extract-tar",
				Wrapped: &fileActions.ExtractFileAction{
					BaseAction:      engine.BaseAction{Logger: logger},
					SourcePath:      "testdata.tar",
					DestinationPath: "extracted-tar",
					ArchiveType:     fileActions.TarArchive,
				},
			},
			// Create a zip archive
			&engine.Action[*CreateArchiveAction]{
				ID: "create-zip-archive",
				Wrapped: &CreateArchiveAction{
					BaseAction:  engine.BaseAction{Logger: logger},
					SourceDir:   "testdata",
					DestPath:    "testdata.zip",
					ArchiveType: fileActions.ZipArchive,
				},
			},
			// Extract the zip archive
			&engine.Action[*fileActions.ExtractFileAction]{
				ID: "extract-zip",
				Wrapped: &fileActions.ExtractFileAction{
					BaseAction:      engine.BaseAction{Logger: logger},
					SourcePath:      "testdata.zip",
					DestinationPath: "extracted-zip",
					ArchiveType:     fileActions.ZipArchive,
				},
			},
		},
	}
}

// NewExtractWithDirectoriesTask demonstrates extraction with directory structures
func NewExtractWithDirectoriesTask(logger *slog.Logger) *engine.Task {
	return &engine.Task{
		Name: "extract-with-directories",
		Actions: []engine.ActionWrapper{
			// Create a complex tar archive with directories
			&engine.Action[*CreateComplexTarAction]{
				ID: "create-complex-tar",
				Wrapped: &CreateComplexTarAction{
					BaseAction: engine.BaseAction{Logger: logger},
					DestPath:   "complex-data.tar",
				},
			},
			// Extract the complex tar archive
			&engine.Action[*fileActions.ExtractFileAction]{
				ID: "extract-complex-tar",
				Wrapped: &fileActions.ExtractFileAction{
					BaseAction:      engine.BaseAction{Logger: logger},
					SourcePath:      "complex-data.tar",
					DestinationPath: "extracted-complex",
					ArchiveType:     fileActions.TarArchive,
				},
			},
		},
	}
}

// NewExtractCompressedArchivesTask demonstrates the composition pattern for compressed archives
func NewExtractCompressedArchivesTask(logger *slog.Logger) *engine.Task {
	return &engine.Task{
		Name: "extract-compressed-archives",
		Actions: []engine.ActionWrapper{
			// Create a tar.gz archive (tar + gzip)
			&engine.Action[*CreateArchiveAction]{
				ID: "create-tar-for-compression",
				Wrapped: &CreateArchiveAction{
					BaseAction:  engine.BaseAction{Logger: logger},
					SourceDir:   "testdata",
					DestPath:    "testdata.tar",
					ArchiveType: fileActions.TarArchive,
				},
			},
			// Compress the tar file with gzip
			&engine.Action[*fileActions.CompressFileAction]{
				ID: "compress-tar",
				Wrapped: &fileActions.CompressFileAction{
					BaseAction:      engine.BaseAction{Logger: logger},
					SourcePath:      "testdata.tar",
					DestinationPath: "testdata.tar.gz",
					CompressionType: fileActions.GzipCompression,
				},
			},
			// Step 1: Decompress the .tar.gz file
			&engine.Action[*fileActions.DecompressFileAction]{
				ID: "decompress-tar-gz",
				Wrapped: &fileActions.DecompressFileAction{
					BaseAction:      engine.BaseAction{Logger: logger},
					SourcePath:      "testdata.tar.gz",
					DestinationPath: "testdata-decompressed.tar",
					CompressionType: fileActions.GzipCompression,
				},
			},
			// Step 2: Extract the decompressed tar file
			&engine.Action[*fileActions.ExtractFileAction]{
				ID: "extract-decompressed-tar",
				Wrapped: &fileActions.ExtractFileAction{
					BaseAction:      engine.BaseAction{Logger: logger},
					SourcePath:      "testdata-decompressed.tar",
					DestinationPath: "extracted-tar-gz",
					ArchiveType:     fileActions.TarArchive,
				},
			},
		},
	}
}

// CreateArchiveAction creates test archives for demonstration
type CreateArchiveAction struct {
	engine.BaseAction
	SourceDir   string
	DestPath    string
	ArchiveType fileActions.ArchiveType
}

func (a CreateArchiveAction) BeforeExecute(ctx context.Context) error {
	// Create test data directory if it doesn't exist
	if err := os.MkdirAll(a.SourceDir, 0750); err != nil {
		return err
	}

	// Create some test files
	testFiles := []string{
		filepath.Join(a.SourceDir, "file1.txt"),
		filepath.Join(a.SourceDir, "file2.txt"),
		filepath.Join(a.SourceDir, "subdir", "file3.txt"),
	}

	for _, file := range testFiles {
		// Create directory if needed
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return err
		}

		// Create test file
		if err := os.WriteFile(file, []byte("Test content for "+file), 0600); err != nil {
			return err
		}
	}

	return nil
}

func (a CreateArchiveAction) Execute(ctx context.Context) error {
	a.Logger.Info("Creating test archive", "source", a.SourceDir, "dest", a.DestPath, "type", a.ArchiveType)

	// Create the archive based on type
	switch a.ArchiveType {
	case fileActions.TarArchive:
		return a.createTarArchive()
	case fileActions.ZipArchive:
		return a.createZipArchive()
	default:
		return fmt.Errorf("unsupported archive type: %s", a.ArchiveType)
	}
}

func (a CreateArchiveAction) AfterExecute(ctx context.Context) error {
	return nil
}

func (a CreateArchiveAction) createTarArchive() error {
	// Create tar file
	tarFile, err := os.Create(a.DestPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	// Walk through the source directory
	return filepath.Walk(a.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(a.SourceDir, path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Create header
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it's a file, write the content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}

func (a CreateArchiveAction) createZipArchive() error {
	// Create zip file
	zipFile, err := os.Create(a.DestPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through the source directory
	return filepath.Walk(a.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(a.SourceDir, path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Create file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Create file in zip
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// If it's a file, write the content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(writer, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// CreateComplexTarAction creates a complex tar archive with nested directories
type CreateComplexTarAction struct {
	engine.BaseAction
	DestPath string
}

func (a CreateComplexTarAction) BeforeExecute(ctx context.Context) error {
	// Create complex test data structure
	testStructure := map[string]string{
		"root.txt":                   "Root file content",
		"dir1/file1.txt":             "File 1 in dir1",
		"dir1/file2.txt":             "File 2 in dir1",
		"dir1/subdir/file3.txt":      "File 3 in subdir",
		"dir2/empty.txt":             "",
		"dir2/nested/deep/file4.txt": "Deep nested file",
		"dir2/nested/deep/file5.txt": "Another deep nested file",
	}

	for path, content := range testStructure {
		fullPath := filepath.Join("testing", "testdata", path)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0750); err != nil {
			return err
		}

		if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
			return err
		}
	}

	return nil
}

func (a CreateComplexTarAction) Execute(ctx context.Context) error {
	a.Logger.Info("Creating complex tar archive", "dest", a.DestPath)

	// Create tar file
	tarFile, err := os.Create(a.DestPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	// Walk through the testdata directory
	return filepath.Walk("testing/testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel("testing/testdata", path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Create header
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it's a file, write the content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}

func (a CreateComplexTarAction) AfterExecute(ctx context.Context) error {
	return nil
}
