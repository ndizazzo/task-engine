# Built-in Actions Inventory

This document provides a comprehensive inventory of all built-in actions available in the Task Engine.

## File & Directory Operations

### CreateDirectoriesAction

Creates multiple directories with proper path handling.

**Constructor:** `NewCreateDirectoriesAction(rootPath string, directories []string, logger *slog.Logger)`

**Parameters:**

- `rootPath`: Base directory path
- `directories`: List of directory names to create
- `logger`: Logger instance

**Example:**

```go
action := actions.NewCreateDirectoriesAction(
    "/tmp/myproject",
    []string{"src", "docs", "tests"},
    logger,
)
```

### WriteFileAction

Writes content to files with optional buffering and overwrite control.

**Constructor:** `NewWriteFileAction(filePath string, content []byte, overwrite bool, inputBuffer *bytes.Buffer, logger *slog.Logger)`

**Parameters:**

- `filePath`: Target file path
- `content`: Content to write (can be nil if using inputBuffer)
- `overwrite`: Whether to overwrite existing files
- `inputBuffer`: Alternative content source (can be nil)
- `logger`: Logger instance

**Example:**

```go
action := actions.NewWriteFileAction(
    "/tmp/config.yaml",
    []byte("key: value\n"),
    true,
    nil,
    logger,
)
```

### CopyFileAction

Copies files with optional directory creation.

**Constructor:** `NewCopyFileAction(source string, destination string, createDirs bool, logger *slog.Logger)`

**Parameters:**

- `source`: Source file path
- `destination`: Destination file path
- `createDirs`: Whether to create destination directories
- `logger`: Logger instance

**Example:**

```go
action := actions.NewCopyFileAction(
    "/tmp/source.txt",
    "/tmp/backup/source.txt",
    true,
    logger,
)
```

### DeleteFileAction

Safely deletes files with proper error handling.

**Constructor:** `NewDeleteFileAction(filePath string, logger *slog.Logger)`

**Parameters:**

- `filePath`: Path to file to delete
- `logger`: Logger instance

**Example:**

```go
action := actions.NewDeleteFileAction("/tmp/tempfile.txt", logger)
```

### ReplaceLinesAction

Replaces text in files using regex patterns.

**Constructor:** `NewReplaceLinesAction(filePath string, pattern string, replacement string, logger *slog.Logger)`

**Parameters:**

- `filePath`: Target file path
- `pattern`: Regex pattern to match
- `replacement`: Replacement text
- `logger`: Logger instance

**Example:**

```go
action := actions.NewReplaceLinesAction(
    "/etc/config.conf",
    `^port=.*`,
    "port=8080",
    logger,
)
```

### ChangeOwnershipAction

Changes file/directory ownership using chown command.

**Constructor:** `NewChangeOwnershipAction(path string, owner string, group string, recursive bool, logger *slog.Logger)`

**Parameters:**

- `path`: Path to file/directory
- `owner`: New owner (can be empty)
- `group`: New group (can be empty)
- `recursive`: Whether to apply recursively
- `logger`: Logger instance

**Example:**

```go
action := actions.NewChangeOwnershipAction(
    "/var/www/html",
    "www-data",
    "www-data",
    true,
    logger,
)
```

### ChangePermissionsAction

Changes file/directory permissions using chmod command.

**Constructor:** `NewChangePermissionsAction(path string, permissions string, recursive bool, logger *slog.Logger)`

**Parameters:**

- `path`: Path to file/directory
- `permissions`: Permissions (octal like "755" or symbolic like "u+x")
- `recursive`: Whether to apply recursively
- `logger`: Logger instance

**Example:**

```go
action := actions.NewChangePermissionsAction(
    "/usr/local/bin/myapp",
    "755",
    false,
    logger,
)
```

### MoveFileAction

Moves/renames files and directories using mv command.

**Constructor:** `NewMoveFileAction(source string, destination string, createDirs bool, logger *slog.Logger)`

**Parameters:**

- `source`: Source path
- `destination`: Destination path
- `createDirs`: Whether to create destination directories
- `logger`: Logger instance

**Example:**

```go
action := actions.NewMoveFileAction(
    "/tmp/oldname.txt",
    "/tmp/newname.txt",
    false,
    logger,
)
```

## Docker Operations

### DockerComposeUpAction

Starts Docker Compose services with optional working directory.

**Constructor:** `NewDockerComposeUpAction(services []string, workingDir string, logger *slog.Logger)`

**Parameters:**

- `services`: List of services to start (empty for all)
- `workingDir`: Working directory for docker-compose command
- `logger`: Logger instance

**Example:**

```go
action := actions.NewDockerComposeUpAction(
    []string{"web", "db"},
    "/path/to/docker-compose",
    logger,
)
```

### DockerComposeDownAction

Stops Docker Compose services with optional working directory.

**Constructor:** `NewDockerComposeDownAction(services []string, workingDir string, logger *slog.Logger)`

**Parameters:**

- `services`: List of services to stop (empty for all)
- `workingDir`: Working directory for docker-compose command
- `logger`: Logger instance

**Example:**

```go
action := actions.NewDockerComposeDownAction(
    []string{},
    "/path/to/docker-compose",
    logger,
)
```

### DockerComposeExecAction

Executes commands in Docker Compose containers.

**Constructor:** `NewDockerComposeExecAction(service string, command []string, workingDir string, logger *slog.Logger)`

**Parameters:**

- `service`: Target service name
- `command`: Command to execute
- `workingDir`: Working directory for docker-compose command
- `logger`: Logger instance

**Example:**

```go
action := actions.NewDockerComposeExecAction(
    "web",
    []string{"php", "artisan", "migrate"},
    "/path/to/project",
    logger,
)
```

### DockerRunAction

Runs Docker containers with flexible configuration.

**Constructor:** `NewDockerRunAction(image string, command []string, options []string, inputBuffer *bytes.Buffer, logger *slog.Logger)`

**Parameters:**

- `image`: Docker image name
- `command`: Command to run in container
- `options`: Docker run options
- `inputBuffer`: Input buffer for container
- `logger`: Logger instance

**Example:**

```go
action := actions.NewDockerRunAction(
    "nginx:alpine",
    []string{"nginx", "-g", "daemon off;"},
    []string{"-p", "80:80", "-d"},
    nil,
    logger,
)
```

### CheckContainerHealthAction

Performs health checks on containers with retry logic.

**Constructor:** `NewCheckContainerHealthAction(containerName string, maxRetries int, retryDelay time.Duration, workingDir string, logger *slog.Logger)`

**Parameters:**

- `containerName`: Name of container to check
- `maxRetries`: Maximum number of retry attempts
- `retryDelay`: Delay between retries
- `workingDir`: Working directory for docker commands
- `logger`: Logger instance

**Example:**

```go
action := actions.NewCheckContainerHealthAction(
    "web-container",
    10,
    5*time.Second,
    "/path/to/project",
    logger,
)
```

### DockerGenericAction

Executes generic Docker commands with flexible arguments.

**Constructor:** `NewDockerGenericAction(args []string, logger *slog.Logger)`

**Parameters:**

- `args`: Docker command arguments
- `logger`: Logger instance

**Example:**

```go
action := actions.NewDockerGenericAction(
    []string{"network", "create", "mynetwork"},
    logger,
)
```

## System Management

### ManageServiceAction

Controls systemd services (start/stop/restart).

**Constructor:** `NewManageServiceAction(serviceName string, action string, logger *slog.Logger)`

**Parameters:**

- `serviceName`: Name of the systemd service
- `action`: Action to perform ("start", "stop", "restart")
- `logger`: Logger instance

**Example:**

```go
action := actions.NewManageServiceAction(
    "nginx",
    "restart",
    logger,
)
```

### ShutdownAction

Performs system shutdown or restart with optional delays.

**Constructor:** `NewShutdownAction(shutdownType string, delay string, logger *slog.Logger)`

**Parameters:**

- `shutdownType`: Type of shutdown ("shutdown", "restart")
- `delay`: Delay before shutdown (e.g., "5", "now")
- `logger`: Logger instance

**Example:**

```go
action := actions.NewShutdownAction(
    "restart",
    "5",
    logger,
)
```

## Utilities

### WaitAction

Waits for a specified duration with context cancellation support.

**Constructor:** `NewWaitAction(duration time.Duration, logger *slog.Logger)`

**Parameters:**

- `duration`: Duration to wait
- `logger`: Logger instance

**Example:**

```go
action := actions.NewWaitAction(
    30*time.Second,
    logger,
)
```

### PrerequisiteCheckAction

Performs conditional execution based on custom check functions.

**Constructor:** `NewPrerequisiteCheckAction(checkFunc func(context.Context) (bool, error), logger *slog.Logger)`

**Parameters:**

- `checkFunc`: Function that returns true if prerequisites are met
- `logger`: Logger instance

**Example:**

```go
action := actions.NewPrerequisiteCheckAction(
    func(ctx context.Context) (bool, error) {
        _, err := os.Stat("/path/to/required/file")
        return err == nil, nil
    },
    logger,
)
```

### FetchInterfacesAction

Retrieves network interface information from the system.

**Constructor:** `NewFetchInterfacesAction(logger *slog.Logger)`

**Parameters:**

- `logger`: Logger instance

**Example:**

```go
action := actions.NewFetchInterfacesAction(logger)
```

## Action Categories Summary

- **File Operations**: 8 actions for comprehensive file/directory management
- **Docker Operations**: 6 actions for container orchestration and management
- **System Management**: 2 actions for system-level operations
- **Utilities**: 3 actions for workflow control and system information

## Common Patterns

### Error Handling

All actions implement proper error handling with:

- Input validation
- Context cancellation support
- Detailed error messages with context
- Structured logging

### Testing Support

All actions support:

- Dependency injection for command runners
- Mock implementations for testing
- Comprehensive test coverage

### Logging

All actions provide:

- Structured logging with consistent fields
- Info-level logging for successful operations
- Error-level logging with detailed context
- Debug-level logging for troubleshooting
