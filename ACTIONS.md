# Built-in Actions Inventory

This document provides a comprehensive inventory of all built-in actions available in the Task Engine. For practical examples and usage patterns, see the example tasks in the `tasks/` directory.

## Action Categories Summary

- **File Operations**: 11 actions for comprehensive file/directory management
- **Docker Operations**: 6 actions for container orchestration and management
- **System Management**: 3 actions for system-level operations
- **Utilities**: 4 actions for workflow control and system information

## File & Directory Operations

### CreateDirectoriesAction

Creates multiple directories with proper path handling.

**Constructor:** `NewCreateDirectoriesAction(rootPath string, directories []string, logger *slog.Logger)`

**Parameters:**

- `rootPath`: Base directory path
- `directories`: List of directory names to create
- `logger`: Logger instance

**See Example:** `tasks.NewFileOperationsTask()` - Demonstrates directory creation as part of a complete file operations workflow.

### WriteFileAction

Writes content to files with optional buffering and overwrite control.

**Constructor:** `NewWriteFileAction(filePath string, content []byte, overwrite bool, inputBuffer *bytes.Buffer, logger *slog.Logger)`

**Parameters:**

- `filePath`: Target file path
- `content`: Content to write (can be nil if using inputBuffer)
- `overwrite`: Whether to overwrite existing files
- `inputBuffer`: Alternative content source (can be nil)
- `logger`: Logger instance

**See Example:** `tasks.NewFileOperationsTask()` - Shows file creation with various content types and configurations.

### ReadFileAction

Reads file contents and stores them in a provided byte array buffer.

**Constructor:** `NewReadFileAction(filePath string, outputBuffer *[]byte, logger *slog.Logger)`

**Parameters:**

- `filePath`: Path to the file to read
- `outputBuffer`: Pointer to byte array where file contents will be stored
- `logger`: Logger instance

**Features:**

- Validates file existence before reading
- Ensures the path is a regular file (not a directory)
- Handles permission errors gracefully
- Supports reading files of any size (limited by available memory)
- Preserves all file content including special characters and unicode

**See Example:** `tasks.NewReadFileOperationsTask()` - Demonstrates file reading with error handling and content processing.

### CompressFileAction

Compresses a file using the specified compression algorithm (currently supports gzip).

**Constructor:** `NewCompressFileAction(sourcePath string, destinationPath string, compressionType CompressionType, logger *slog.Logger)`

**Parameters:**

- `sourcePath`: Source file path to compress
- `destinationPath`: Destination path for the compressed file
- `compressionType`: Type of compression to use (e.g., `file.GzipCompression`)
- `logger`: Logger instance

**Supported Compression Types:**

- `file.GzipCompression`: Gzip compression (`.gz` files)

**Features:**

- Validates source file existence and type
- Creates destination directories automatically
- Provides compression ratio information
- Handles large files efficiently
- Supports empty files

**See Example:** `tasks.NewCompressionOperationsTask()` - Shows compression workflows with multiple files and auto-detection.

### DecompressFileAction

Decompresses a file using the specified compression algorithm. Supports auto-detection from file extension.

**Constructor:** `NewDecompressFileAction(sourcePath string, destinationPath string, compressionType CompressionType, logger *slog.Logger)`

**Parameters:**

- `sourcePath`: Source compressed file path
- `destinationPath`: Destination path for the decompressed file
- `compressionType`: Type of compression (can be empty for auto-detection)
- `logger`: Logger instance

**Auto-Detection:**

When `compressionType` is empty, the action will auto-detect the compression type from the file extension:

- `.gz`, `.gzip` → Gzip compression

**Features:**

- Validates source file existence and type
- Creates destination directories automatically
- Provides compression ratio information
- Supports auto-detection from file extensions
- Handles large files efficiently

**See Example:** `tasks.NewCompressionWithAutoDetectTask()` - Demonstrates auto-detection capabilities.

### CopyFileAction

Copies files and directories with optional recursive copying and directory creation.

**Constructor:** `NewCopyFileAction(source string, destination string, createDirs bool, recursive bool, logger *slog.Logger)`

**Parameters:**

- `source`: Source file or directory path
- `destination`: Destination file or directory path
- `createDirs`: Whether to create destination directories
- `recursive`: Whether to copy directories recursively (uses -R flag equivalent)
- `logger`: Logger instance

**See Example:** `tasks.NewFileOperationsTask()` - Shows file copying as part of backup and workflow operations.

### DeletePathAction

Safely deletes files and directories with optional recursive deletion and dry-run support.

**Constructor:** `NewDeletePathAction(path string, recursive bool, dryRun bool, logger *slog.Logger)`

**Parameters:**

- `path`: Path to file or directory to delete
- `recursive`: Whether to delete directories recursively
- `dryRun`: Whether to simulate deletion without actually deleting
- `logger`: Logger instance

**Features:**

- Supports both files and directories
- Recursive directory deletion with comprehensive logging
- Dry-run mode for safe testing
- Detailed deletion planning and statistics
- Handles symlinks and special files
- Graceful handling of non-existent paths

**See Example:** `tasks.NewFileOperationsTask()` - Demonstrates safe file deletion and cleanup operations.

### ReplaceLinesAction

Replaces text in files using regex patterns.

**Constructor:** `NewReplaceLinesAction(filePath string, pattern string, replacement string, logger *slog.Logger)`

**Parameters:**

- `filePath`: Target file path
- `pattern`: Regex pattern to match
- `replacement`: Replacement text
- `logger`: Logger instance

**See Example:** `tasks.NewFileOperationsTask()` - Shows text replacement in configuration files and source code.

### ChangeOwnershipAction

Changes file/directory ownership using chown command.

**Constructor:** `NewChangeOwnershipAction(path string, owner string, group string, recursive bool, logger *slog.Logger)`

**Parameters:**

- `path`: Path to file/directory
- `owner`: New owner (can be empty)
- `group`: New group (can be empty)
- `recursive`: Whether to apply recursively
- `logger`: Logger instance

### ChangePermissionsAction

Changes file/directory permissions using chmod command.

**Constructor:** `NewChangePermissionsAction(path string, permissions string, recursive bool, logger *slog.Logger)`

**Parameters:**

- `path`: Path to file/directory
- `permissions`: Permissions (octal like "755" or symbolic like "u+x")
- `recursive`: Whether to apply recursively
- `logger`: Logger instance

### MoveFileAction

Moves/renames files and directories using mv command.

**Constructor:** `NewMoveFileAction(source string, destination string, createDirs bool, logger *slog.Logger)`

**Parameters:**

- `source`: Source path
- `destination`: Destination path
- `createDirs`: Whether to create destination directories
- `logger`: Logger instance

## Docker Operations

### DockerComposeUpAction

Starts Docker Compose services with optional working directory.

**Constructor:** `NewDockerComposeUpAction(services []string, workingDir string, logger *slog.Logger)`

**Parameters:**

- `services`: List of services to start (empty for all)
- `workingDir`: Working directory for docker-compose command
- `logger`: Logger instance

**See Example:** `tasks.NewDockerSetupTask()` - Demonstrates Docker environment setup and service management.

### DockerComposeDownAction

Stops Docker Compose services with optional working directory.

**Constructor:** `NewDockerComposeDownAction(services []string, workingDir string, logger *slog.Logger)`

**Parameters:**

- `services`: List of services to stop (empty for all)
- `workingDir`: Working directory for docker-compose command
- `logger`: Logger instance

### DockerComposeExecAction

Executes commands in Docker Compose containers.

**Constructor:** `NewDockerComposeExecAction(service string, command []string, workingDir string, logger *slog.Logger)`

**Parameters:**

- `service`: Target service name
- `command`: Command to execute
- `workingDir`: Working directory for docker-compose command
- `logger`: Logger instance

### DockerRunAction

Runs Docker containers with flexible configuration.

**Constructor:** `NewDockerRunAction(image string, command []string, options []string, inputBuffer *bytes.Buffer, logger *slog.Logger)`

**Parameters:**

- `image`: Docker image name
- `command`: Command to run in container
- `options`: Docker run options
- `inputBuffer`: Input buffer for container
- `logger`: Logger instance

### CheckContainerHealthAction

Performs health checks on containers with retry logic.

**Constructor:** `NewCheckContainerHealthAction(containerName string, maxRetries int, retryDelay time.Duration, workingDir string, logger *slog.Logger)`

**Parameters:**

- `containerName`: Name of container to check
- `maxRetries`: Maximum number of retry attempts
- `retryDelay`: Delay between retries
- `workingDir`: Working directory for docker commands
- `logger`: Logger instance

### DockerGenericAction

Executes generic Docker commands with flexible arguments.

**Constructor:** `NewDockerGenericAction(args []string, logger *slog.Logger)`

**Parameters:**

- `args`: Docker command arguments
- `logger`: Logger instance

## System Management

### ManageServiceAction

Controls systemd services (start/stop/restart).

**Constructor:** `NewManageServiceAction(serviceName string, action string, logger *slog.Logger)`

**Parameters:**

- `serviceName`: Name of the systemd service
- `action`: Action to perform ("start", "stop", "restart")
- `logger`: Logger instance

**See Example:** `tasks.NewSystemManagementTask()` - Demonstrates system service management operations.

### ShutdownAction

Performs system shutdown or restart with optional delays.

**Constructor:** `NewShutdownAction(shutdownType string, delay string, logger *slog.Logger)`

**Parameters:**

- `shutdownType`: Type of shutdown ("shutdown", "restart")
- `delay`: Delay before shutdown (e.g., "5", "now")
- `logger`: Logger instance

### UpdatePackagesAction

Installs packages using the appropriate package manager for the operating system. Supports Debian-based Linux (apt) and macOS (Homebrew).

**Constructor:** `NewUpdatePackagesAction(packageNames []string, logger *slog.Logger)`

**Parameters:**

- `packageNames`: List of package names to install
- `logger`: Logger instance

**Supported Platforms:**

- **Debian-based Linux**: Uses `apt install -y` (automatically updates package list first)
- **macOS**: Uses `brew install`
- **Auto-detection**: Automatically detects the appropriate package manager

**Features:**

- Non-interactive installation (uses `-y` flag for apt)
- Automatic package list updates for apt
- Cross-platform compatibility
- Comprehensive error handling
- Detailed logging of installation progress

**See Example:** `tasks.NewPackageManagementTask()` - Shows package installation workflows for different platforms.

## Utilities

### WaitAction

Waits for a specified duration with context cancellation support.

**Constructor:** `NewWaitAction(duration time.Duration, logger *slog.Logger)`

**Parameters:**

- `duration`: Duration to wait
- `logger`: Logger instance

**See Example:** `tasks.NewUtilityOperationsTask()` - Demonstrates utility operations including timing and delays.

### PrerequisiteCheckAction

Performs conditional execution based on custom check functions.

**Constructor:** `NewPrerequisiteCheckAction(checkFunc func(context.Context) (bool, error), logger *slog.Logger)`

**Parameters:**

- `checkFunc`: Function that returns true if prerequisites are met
- `logger`: Logger instance

### FetchInterfacesAction

Retrieves network interface information from the system.

**Constructor:** `NewFetchInterfacesAction(logger *slog.Logger)`

**Parameters:**

- `logger`: Logger instance

### ReadMACAddressAction

Reads the MAC address of a specific network interface from the system.

**Constructor:** `NewReadMacAction(netInterface string, logger *slog.Logger)`

**Parameters:**

- `netInterface`: Name of the network interface (e.g., "eth0", "wlan0")
- `logger`: Logger instance

**Features:**

- Reads MAC address from `/sys/class/net/{interface}/address`
- Trims whitespace from the result
- Stores MAC address in the action for later retrieval

## Example Tasks

For practical examples and complete workflows, see the following task functions in the `tasks/` directory:

- **File Operations**: `tasks.NewFileOperationsTask()` - Complete file management workflow
- **Compression**: `tasks.NewCompressionOperationsTask()` - File compression and decompression
- **Package Management**: `tasks.NewPackageManagementTask()` - Cross-platform package installation
- **System Management**: `tasks.NewSystemManagementTask()` - System-level operations
- **Utility Operations**: `tasks.NewUtilityOperationsTask()` - Utility and helper operations
- **Docker Setup**: `tasks.NewDockerSetupTask()` - Docker environment configuration

Each example task demonstrates real-world usage patterns and can be used as a starting point for your own workflows.
