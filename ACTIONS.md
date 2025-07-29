# Built-in Actions Inventory

This document provides a comprehensive inventory of all built-in actions available in the Task Engine. For practical examples and usage patterns, see the example tasks in the `tasks/` directory.

## Action Categories Summary

- **File Operations**: 13 actions for comprehensive file/directory management
- **Docker Operations**: 14 actions for container orchestration and management
- **System Management**: 4 actions for system-level operations
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

### CreateSymlinkAction

Creates symbolic links with comprehensive validation and verification.

**Constructor:** `NewCreateSymlinkAction(target string, linkPath string, overwrite bool, createDirs bool, logger *slog.Logger)`

**Parameters:**

- `target`: Target file or directory path for the symlink
- `linkPath`: Path where the symlink will be created
- `overwrite`: Whether to overwrite existing symlinks
- `createDirs`: Whether to create parent directories for the symlink
- `logger`: Logger instance

**Features:**

- **Path Validation**: Validates both target and link paths to prevent path traversal attacks
- **Overwrite Control**: Optionally overwrites existing symlinks with proper cleanup
- **Directory Creation**: Automatically creates parent directories when requested
- **Comprehensive Verification**: Verifies symlink creation with target validation
- **Relative Path Support**: Handles both absolute and relative target paths
- **Error Handling**: Detailed error messages for various failure scenarios
- **Security**: Path sanitization to prevent malicious path traversal

**Verification Process:**

The action includes a three-step verification process:
1. **Existence Check**: Verifies the symlink exists and is actually a symlink
2. **Target Reading**: Reads the symlink target to ensure it's accessible
3. **Target Comparison**: Compares the actual target with the expected target, handling both absolute and relative paths

**Error Scenarios:**

- Invalid target or link paths (empty, path traversal attempts)
- Target and link path being the same
- Permission errors when creating symlinks or directories
- Existing symlinks when overwrite is disabled
- Verification failures (not a symlink, target mismatch, broken symlinks)

**See Example:** `tasks.NewSymlinkOperationsTask()` - Demonstrates various symlink creation scenarios including relative paths, overwrite operations, and directory creation.

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

### ExtractFileAction

Extracts files from archive formats (tar, zip) with comprehensive security features and auto-detection.

**Constructor:** `NewExtractFileAction(sourcePath string, destinationPath string, archiveType ArchiveType, logger *slog.Logger)`

**Parameters:**

- `sourcePath`: Path to the archive file to extract
- `destinationPath`: Directory where files will be extracted
- `archiveType`: Type of archive (can be empty for auto-detection)
- `logger`: Logger instance

**Supported Archive Types:**

- `file.TarArchive`: Tar archives (`.tar`)
- `file.ZipArchive`: Zip archives (`.zip`)
- `file.AutoDetect`: Auto-detect from file extension

**Auto-Detection:**
When `archiveType` is `file.AutoDetect`, the action will auto-detect the archive type from the file extension:

- `.tar` → Tar archive
- `.zip` → Zip archive

**Security Features:**

- **Path Traversal Protection**: Validates and sanitizes file paths to prevent zip slip attacks
- **Decompression Bomb Protection**: Limits file sizes to prevent memory exhaustion attacks
- **Integer Overflow Protection**: Safe permission setting with bit masking
- **Insecure Permissions Prevention**: Uses secure default permissions (0600 for files, 0750 for directories)

**Features:**

- Validates source file existence and type
- Creates destination directories automatically
- Supports both tar and zip archive formats
- Auto-detection from file extensions
- Comprehensive security measures
- Handles large archives efficiently
- Preserves file permissions (when safe)
- Detailed error reporting

**Error Handling:**

- Returns error if compressed archives (`.tar.gz`) are passed directly
- Validates archive integrity before extraction
- Handles corrupted or invalid archives gracefully
- Provides detailed error messages for troubleshooting

**See Example:** `tasks.NewExtractOperationsTask()` - Demonstrates archive extraction workflows with security considerations.

## Docker Operations

### GetDockerStatusAction

Retrieves the state of specific Docker containers by ID or name.

**Constructor:** `NewGetContainerStateAction(logger *slog.Logger, containerIdentifiers ...string)`

**Parameters:**

- `logger`: Logger instance
- `containerIdentifiers`: Variable number of container IDs or names to query

**Returns:** `ContainerState` struct with ID, names, image, and status information

**Features:**

- Supports multiple container queries in a single action
- Handles containers with multiple names/aliases
- Robust JSON parsing with error recovery
- Returns structured container information

### GetAllContainersStateAction

Retrieves the state of all Docker containers on the system.

**Constructor:** `NewGetAllContainersStateAction(logger *slog.Logger)`

**Parameters:**

- `logger`: Logger instance

**Returns:** Array of `ContainerState` structs for all containers

**Features:**

- Comprehensive container enumeration
- Includes stopped, running, and paused containers
- Efficient JSON parsing for large container lists

### ContainerState Struct

Container state information is returned in a structured format:

```go
type ContainerState struct {
    ID     string   // Container ID
    Names  []string // Container names (can be multiple)
    Image  string   // Container image
    Status string   // Container status with full state information
}
```

**Supported Container States:**

- `created` - Container created but not started
- `restarting` - Container is restarting (with restart policy)
- `running` - Container is running (shows as "Up X time")
- `removing` - Container is being removed
- `paused` - Container is paused
- `exited` - Container has exited (shows as "Exited (code) X time ago")
- `dead` - Container is dead (failed to start or was killed)

**Status Examples:**

- `"Up 2 hours"` - Running for 2 hours
- `"Exited (0) 1 hour ago"` - Exited with code 0, 1 hour ago
- `"Paused"` - Currently paused
- `"Created"` - Created but not started
- `"Restarting (1) 2 minutes ago"` - Restarting, attempt 1, 2 minutes ago
- `"Dead"` - Container is dead
- `"Removing"` - Container is being removed

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

### DockerLoadAction

Loads Docker images from tar archive files with support for platform filtering and quiet mode.

**Constructor:** `NewDockerLoadAction(logger *slog.Logger, tarFilePath string, options ...DockerLoadOption)`

**Parameters:**

- `logger`: Logger instance
- `tarFilePath`: Path to the tar archive file containing Docker images
- `options`: Optional configuration options (see below)

**Options:**

- `WithPlatform(platform string)`: Load only the specified platform variant (e.g., "linux/amd64", "linux/arm64")
- `WithQuiet()`: Suppress the load output

**Features:**

- Loads images from tar archive files using `docker load -i`
- Supports platform-specific loading for multi-architecture images
- Quiet mode for reduced output in automated workflows
- Parses and stores loaded image names and IDs
- Comprehensive error handling with detailed error messages
- Flexible configuration through functional options pattern

**Returns:** Array of loaded image names/IDs in the `LoadedImages` field

**See Example:** `tasks.NewDockerLoadTask()` - Demonstrates various Docker load operations including platform-specific loading and batch operations.

### DockerImageRmAction

Removes Docker images by name/tag or ID with support for force removal and pruning control.

**Constructors:**

- `NewDockerImageRmByNameAction(logger *slog.Logger, imageName string, options ...DockerImageRmOption)`: Remove image by name and tag
- `NewDockerImageRmByIDAction(logger *slog.Logger, imageID string, options ...DockerImageRmOption)`: Remove image by ID

**Parameters:**

- `logger`: Logger instance
- `imageName`: Image name and tag (e.g., "nginx:latest", "my-registry.com/namespace/image:v1.0")
- `imageID`: Image ID (e.g., "sha256:abc123def456789")
- `options`: Optional configuration options (see below)

**Options:**

- `WithForce()`: Force the removal of the image (equivalent to `docker image rm -f`)
- `WithNoPrune()`: Prevent removal of untagged parent images (equivalent to `docker image rm --no-prune`)

**Features:**

- Two distinct constructors for different use cases (name/tag vs ID)
- Supports force removal for images that might be in use
- Controls pruning behavior for parent layers
- Parses and stores removed image names and IDs
- Comprehensive error handling with detailed error messages
- Flexible configuration through functional options pattern
- Handles registry images with complex naming patterns

**Returns:** Array of removed image names/IDs in the `RemovedImages` field

**See Example:** `tasks.NewDockerImageRmTask()` - Demonstrates various Docker image removal operations including force removal and cleanup workflows.

### DockerImageListAction

Lists Docker images with comprehensive metadata including repository, tag, image ID, creation time, and size.

**Constructor:** `NewDockerImageListAction(logger *slog.Logger, options ...DockerImageListOption)`

**Parameters:**

- `logger`: Logger instance
- `options`: Optional configuration options (see below)

**Options:**

- `WithAll()`: Show all images (including intermediate images)
- `WithDigests()`: Show digests
- `WithFilter(filter string)`: Filter output based on conditions
- `WithFormat(format string)`: Use a custom template for output
- `WithNoTrunc()`: Don't truncate output
- `WithQuiet()`: Only show image IDs

**Features:**

- Lists all Docker images with detailed metadata
- Supports filtering by various criteria (dangling, label, before, since, etc.)
- Custom output formatting using Go templates
- Handles dangling images (`<none>:<none>`)
- Parses and stores structured image information
- Comprehensive error handling with detailed error messages
- Flexible configuration through functional options pattern

**Returns:** Array of `DockerImage` structs with repository, tag, image ID, created time, and size

**DockerImage Structure:**

```go
type DockerImage struct {
    Repository string
    Tag        string
    ImageID    string
    Created    string
    Size       string
}
```

**See Example:** `tasks.NewDockerImageListTask()` - Demonstrates various Docker image listing operations including filtering and formatting.

### DockerComposeLsAction

Lists Docker Compose stacks with their status and configuration files.

**Constructor:** `NewDockerComposeLsAction(logger *slog.Logger, options ...DockerComposeLsOption)`

**Parameters:**

- `logger`: Logger instance
- `options`: Optional configuration options (see below)

**Options:**

- `WithAll()`: Show all stacks (including stopped ones)
- `WithFilter(filter string)`: Filter output based on conditions
- `WithFormat(format string)`: Use a custom template for output
- `WithQuiet()`: Only show stack names
- `WithWorkingDir(workingDir string)`: Set working directory for docker-compose command

**Features:**

- Lists all Docker Compose stacks with their current status
- Shows configuration files used by each stack
- Supports filtering by stack name or other criteria
- Custom output formatting using Go templates
- Parses and stores structured stack information
- Comprehensive error handling with detailed error messages
- Flexible configuration through functional options pattern

**Returns:** Array of `ComposeStack` structs with name, status, and config files

**ComposeStack Structure:**

```go
type ComposeStack struct {
    Name        string
    Status      string
    ConfigFiles string
}
```

**See Example:** `tasks.NewDockerComposeLsTask()` - Demonstrates Docker Compose stack listing operations including filtering and status monitoring.

### DockerComposePsAction

Lists services for Docker Compose stacks with detailed container information.

**Constructor:** `NewDockerComposePsAction(services []string, logger *slog.Logger, options ...DockerComposePsOption)`

**Parameters:**

- `services`: List of services to list (empty for all)
- `logger`: Logger instance
- `options`: Optional configuration options (see below)

**Options:**

- `WithAll()`: Show all containers (including stopped ones)
- `WithFilter(filter string)`: Filter output based on conditions
- `WithFormat(format string)`: Use a custom template for output
- `WithQuiet()`: Only show container names
- `WithWorkingDir(workingDir string)`: Set working directory for docker-compose command

**Features:**

- Lists services for Docker Compose stacks with detailed information
- Shows container name, image, command, service name, status, and ports
- Supports filtering by service status or other criteria
- Custom output formatting using Go templates
- Parses and stores structured service information
- Comprehensive error handling with detailed error messages
- Flexible configuration through functional options pattern

**Returns:** Array of `ComposeService` structs with container details

**ComposeService Structure:**

```go
type ComposeService struct {
    Name        string
    Image       string
    Command     string
    ServiceName string
    Status      string
    Ports       string
}
```

**See Example:** `tasks.NewDockerComposePsTask()` - Demonstrates Docker Compose service listing operations including status monitoring and port mapping.

### DockerPsAction

Lists Docker containers with comprehensive metadata including container ID, image, command, status, ports, and names.

**Constructor:** `NewDockerPsAction(logger *slog.Logger, options ...DockerPsOption)`

**Parameters:**

- `logger`: Logger instance
- `options`: Optional configuration options (see below)

**Options:**

- `WithPsAll()`: Show all containers (including stopped ones)
- `WithPsFilter(filter string)`: Filter output based on conditions
- `WithPsFormat(format string)`: Use a custom template for output
- `WithPsLast(n int)`: Show n last created containers
- `WithPsLatest()`: Show the latest created container
- `WithPsNoTrunc()`: Don't truncate output
- `WithPsQuiet()`: Only show container IDs
- `WithPsSize()`: Display total file sizes

**Features:**

- Lists all Docker containers with detailed metadata
- Supports filtering by various criteria (status, label, ancestor, etc.)
- Custom output formatting using Go templates
- Shows container size information when requested
- Parses and stores structured container information
- Comprehensive error handling with detailed error messages
- Flexible configuration through functional options pattern

**Returns:** Array of `Container` structs with container details

**Container Structure:**

```go
type Container struct {
    ContainerID string
    Image       string
    Command     string
    Created     string
    Status      string
    Ports       string
    Names       string
}
```

**See Example:** `tasks.NewDockerPsTask()` - Demonstrates various Docker container listing operations including filtering, formatting, and status monitoring.

### DockerPullAction

Pulls Docker images with support for single architecture specifications and platform options.

**Constructor:** `NewDockerPullAction(logger *slog.Logger, images map[string]ImageSpec, options ...DockerPullOption)`

**Parameters:**

- `logger`: Logger instance
- `images`: Map of image names to `ImageSpec` configurations
- `options`: Optional configuration options (see below)

**ImageSpec Structure:**

```go
type ImageSpec struct {
    Image        string
    Tag          string
    Architecture string
}
```

**Options:**

- `WithAllTags()`: Pull all tags for the specified images
- `WithPullQuietOutput()`: Suppress verbose output
- `WithPullPlatform(platform string)`: Specify platform for pulled images

**Features:**

- Pulls Docker images with architecture-specific platform targeting
- Supports multiple images in a single action
- Platform override options for cross-platform compatibility
- Quiet mode for reduced output in automated workflows
- Comprehensive error handling with detailed error messages
- Tracks successfully pulled and failed images
- Flexible configuration through functional options pattern

**Returns:** Arrays of successfully pulled and failed image names in `PulledImages` and `FailedImages` fields

**See Example:** `tasks.NewDockerPullTask()` - Demonstrates Docker image pulling operations with various configurations.

### DockerPullMultiArchAction

Pulls Docker images for multiple architectures in a single action with robust error handling.

**Constructor:** `NewDockerPullMultiArchAction(logger *slog.Logger, multiArchImages map[string]MultiArchImageSpec, options ...DockerPullOption)`

**Parameters:**

- `logger`: Logger instance
- `multiArchImages`: Map of image names to `MultiArchImageSpec` configurations
- `options`: Optional configuration options (see below)

**MultiArchImageSpec Structure:**

```go
type MultiArchImageSpec struct {
    Image         string
    Tag           string
    Architectures []string
}
```

**Options:**

- `WithAllTags()`: Pull all tags for the specified images
- `WithPullQuietOutput()`: Suppress verbose output
- `WithPullPlatform(platform string)`: Override platform for all architectures

**Features:**

- Pulls the same image for multiple architectures (e.g., amd64, arm64, arm/v7)
- Robust error handling with partial success support
- Continues pulling other architectures even if some fail
- Detailed logging per architecture
- Platform override capability for all architectures
- Tracks successfully pulled and failed images
- Flexible configuration through functional options pattern

**Error Handling:**

- **Partial Success**: If some architectures fail, others still succeed
- **Complete Failure**: Only fails if ALL architectures fail for an image
- **Warning Messages**: Alerts when partial failures occur

**Returns:** Arrays of successfully pulled and failed image names in `PulledImages` and `FailedImages` fields

**See Example:** `tasks.NewDockerPullMultiArchTask()` - Demonstrates multi-architecture Docker image pulling operations.

### DockerGenericAction

Executes generic Docker commands with flexible arguments.

**Constructor:** `NewDockerGenericAction(args []string, logger *slog.Logger)`

**Parameters:**

- `args`: Docker command arguments
- `logger`: Logger instance

## System Management

### GetServiceStatusAction

Retrieves the status of one or more systemd services using `systemctl show` with specific properties for reliable parsing.

**Constructor:** `NewGetServiceStatusAction(logger *slog.Logger, serviceNames ...string)`

**Parameters:**

- `logger`: Logger instance
- `serviceNames`: Variable number of service names to check

**Features:**

- Uses `systemctl show` with specific properties for reliable, machine-readable output
- Handles multiple services in a single action
- Retrieves comprehensive service information including:
  - Service description and documentation
  - Loaded state and vendor information
  - Active state and sub-status
  - Service path and configuration details
- Gracefully handles non-existent services
- More efficient and reliable than parsing human-readable `systemctl status` output
- Handles various service states (running, exited, dead, etc.)
- Processes each service individually to avoid mixed output issues

**ServiceStatus Structure:**

```go
type ServiceStatus struct {
    Name        string `json:"name"`
    Loaded      string `json:"loaded"`
    Active      string `json:"active"`
    Sub         string `json:"sub"`
    Description string `json:"description"`
    Path        string `json:"path"`
    Vendor      string `json:"vendor"`
    Exists      bool   `json:"exists"`
}
```

**See Example:** `tasks.NewServiceStatusTask()` - Demonstrates service status checking and health monitoring.

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
- **Container State**: `tasks.NewContainerStateTask()` - Container monitoring and state management
- **Docker Pull**: `tasks.NewDockerPullTask()` - Docker image pulling operations
- **Docker Pull Multi-Arch**: `tasks.NewDockerPullMultiArchTask()` - Multi-architecture image pulling
- **Extract Operations**: `tasks.NewExtractOperationsTask()` - Archive extraction with security features
- **Symlink Operations**: `tasks.NewSymlinkOperationsTask()` - Symbolic link creation and management

Each example task demonstrates real-world usage patterns and can be used as a starting point for your own workflows.
