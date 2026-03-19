# Built-in Actions

Complete list of available actions. See `tasks/` directory for usage examples.

## File Operations

### CreateDirectoriesAction

Creates multiple directories.

```go
action, err := file.NewCreateDirectoriesAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/root"},
    engine.StaticParameter{Value: []string{"src", "tests", "docs"}},
)
```

### WriteFileAction

Writes content to files.

```go
action, err := file.NewWriteFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/file"},
    engine.StaticParameter{Value: []byte("content")},
    true,  // overwrite
    nil,   // inputBuffer (use content param instead)
)
```

### ReadFileAction

Reads file contents into a buffer.

```go
var content []byte
action, err := file.NewReadFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/file"},
    &content,
)
```

### CompressFileAction

Compresses files (gzip).

```go
action, err := file.NewCompressFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/source/file"},
    engine.StaticParameter{Value: "/dest/file.gz"},
    file.GzipCompression,
)
```

### DecompressFileAction

Decompresses files with auto-detection.

```go
action, err := file.NewDecompressFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/source/file.gz"},
    engine.StaticParameter{Value: "/dest/file"},
    "", // auto-detect
)
```

### CopyFileAction

Copies files and directories.

```go
action, err := file.NewCopyFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/source"},
    engine.StaticParameter{Value: "/dest"},
    true,  // createDirs
    false, // recursive
)
```

### DeletePathAction

Safely deletes files and directories.

```go
action, err := file.NewDeletePathAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/delete"},
    true,  // recursive
    false, // dryRun
    false, // includeHidden
    nil,   // excludePatterns
)
```

### ReplaceLinesAction

Replaces text using regex patterns.

```go
action, err := file.NewReplaceLinesAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/file"},
    map[*regexp.Regexp]engine.ActionParameter{
        regexp.MustCompile("old"): engine.StaticParameter{Value: "new"},
    },
)
```

### CreateSymlinkAction

Creates symbolic links.

```go
action, err := file.NewCreateSymlinkAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/target"},
    engine.StaticParameter{Value: "/path/to/link"},
    false, // overwrite existing link
    true,  // createDirs (create parent directories)
)
```

### ChangeOwnershipAction

Changes file ownership.

```go
action, err := file.NewChangeOwnershipAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path"},
    engine.StaticParameter{Value: "user"},
    engine.StaticParameter{Value: "group"},
    true, // recursive
)
```

### ChangePermissionsAction

Changes file permissions.

```go
action, err := file.NewChangePermissionsAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path"},
    engine.StaticParameter{Value: "755"},
    true, // recursive
)
```

### MoveFileAction

Moves/renames files.

```go
action, err := file.NewMoveFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/source"},
    engine.StaticParameter{Value: "/dest"},
    true, // createDirs
)
```

### ExtractFileAction

Extracts archives (tar, zip) with security features.

```go
action, err := file.NewExtractFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/archive.tar"},
    engine.StaticParameter{Value: "/extract/dir"},
    file.AutoDetect, // or file.TarArchive, file.ZipArchive
)
```

## Docker Operations

### GetContainerStateAction

Gets container status by name or ID. Pass multiple names/IDs via a `[]string`; omit the parameter (pass `nil`) to get all containers.

```go
// Get specific containers
action, err := docker.NewGetContainerStateAction(logger).WithParameters(
    engine.StaticParameter{Value: []string{"container1", "container2"}},
)

// Get all containers
action, err := docker.NewGetContainerStateAction(logger).WithParameters(nil)
```

### DockerComposeUpAction

Starts Docker Compose services.

```go
action, err := docker.NewDockerComposeUpAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/compose"}, // workingDir
    engine.StaticParameter{Value: []string{"web", "db"}}, // services
)
```

### DockerComposeDownAction

Stops Docker Compose services.

```go
action, err := docker.NewDockerComposeDownAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/compose"},
    engine.StaticParameter{Value: []string{"web"}},
)
```

### DockerComposeExecAction

Executes commands in containers.

```go
action, err := docker.NewDockerComposeExecAction(logger).WithParameters(
    engine.StaticParameter{Value: "web"},
    engine.StaticParameter{Value: []string{"ls", "-la"}},
    engine.StaticParameter{Value: "/path/to/compose"},
)
```

### DockerRunAction

Runs a Docker container. Accepts an image parameter, an optional output buffer, and variadic run arguments.

```go
action, err := docker.NewDockerRunAction(logger).WithParameters(
    engine.StaticParameter{Value: "nginx:latest"}, // image
    nil,                                           // outputBuffer (*bytes.Buffer)
    "-p", "8080:80", "-d",                         // runArgs (variadic)
)

// With dynamic image from a previous action/task:
action, err := docker.NewDockerRunAction(logger).WithParameters(
    engine.TaskOutputField("build-task", "imageID"),
    nil,
    "-p", "8080:80", "-d",
)
```

### CheckContainerHealthAction

Performs health checks with retries.

```go
action, err := docker.NewCheckContainerHealthAction(logger).WithParameters(
    engine.StaticParameter{Value: "container"},
    engine.StaticParameter{Value: 3},            // maxRetries
    engine.StaticParameter{Value: time.Second},  // retryDelay
    engine.StaticParameter{Value: "/workdir"},
)
```

### DockerLoadAction

Loads images from tar files.

```go
docker.NewDockerLoadAction(logger, "/path/to/image.tar")
// With options:
docker.NewDockerLoadAction(logger, "/path/to/image.tar",
    docker.WithPlatform("linux/amd64"),
    docker.WithQuiet(),
)
```

### DockerImageRmAction

Removes Docker images.

```go
// By name
docker.NewDockerImageRmByNameAction(logger, "nginx:latest")
// By ID
docker.NewDockerImageRmByIDAction(logger, "sha256:abc123")
// With options:
docker.NewDockerImageRmByNameAction(logger, "nginx:latest",
    docker.WithForce(),
    docker.WithNoPrune(),
)
```

### DockerImageListAction

Lists Docker images.

```go
docker.NewDockerImageListAction(logger)
// With options:
docker.NewDockerImageListAction(logger,
    docker.WithAll(),
    docker.WithDigests(),
    docker.WithFilter("dangling=true"),
)
```

### DockerComposeLsAction

Lists Docker Compose stacks.

```go
docker.NewDockerComposeLsAction(logger)
// With options:
docker.NewDockerComposeLsAction(logger,
    docker.WithAll(),
    docker.WithWorkingDir("/path/to/compose"),
)
```

### DockerComposePsAction

Lists services in Docker Compose stacks.

```go
docker.NewDockerComposePsAction(logger, []string{"web"})
// With options:
docker.NewDockerComposePsAction(logger, []string{"web"},
    docker.WithAll(),
    docker.WithWorkingDir("/path/to/compose"),
)
```

### DockerPsAction

Lists Docker containers.

```go
docker.NewDockerPsAction(logger)
// With options:
docker.NewDockerPsAction(logger,
    docker.WithPsAll(),
    docker.WithPsFilter("status=running"),
)
```

### DockerPullAction

Pulls Docker images.

```go
images := map[string]docker.ImageSpec{
    "nginx": {Image: "nginx", Tag: "latest", Architecture: "amd64"},
}
docker.NewDockerPullAction(logger, images)
// With options:
docker.NewDockerPullAction(logger, images,
    docker.WithAllTags(),
    docker.WithPullPlatform("linux/amd64"),
)
```

### DockerPullMultiArchAction

Pulls images for multiple architectures.

```go
images := map[string]docker.MultiArchImageSpec{
    "nginx": {
        Image: "nginx",
        Tag: "latest",
        Architectures: []string{"amd64", "arm64"},
    },
}
docker.NewDockerPullMultiArchAction(logger, images)
```

### DockerGenericAction

Executes generic Docker commands.

```go
docker.NewDockerGenericAction(logger, []string{"images", "-q"})
```

## System Management

### ServiceStatusAction

Gets systemd service status for one or more services.

```go
action, err := system.NewServiceStatusAction(logger).WithParameters(
    engine.StaticParameter{Value: []string{"nginx", "mysql"}},
)
```

### ManageServiceAction

Controls systemd services (start, stop, restart, enable, disable).

```go
action, err := system.NewManageServiceAction(logger).WithParameters(
    engine.StaticParameter{Value: "nginx"},
    engine.StaticParameter{Value: "restart"},
)
```

### ShutdownAction

Shuts down or restarts system.

```go
action, err := system.NewShutdownAction(logger).WithParameters(
    engine.StaticParameter{Value: "restart"},
    engine.StaticParameter{Value: "5"}, // delay in minutes
)
```

### UpdatePackagesAction

Installs packages (apt/brew). Package manager is auto-detected from the OS; pass an empty string to auto-detect.

```go
action, err := system.NewUpdatePackagesAction(logger).WithParameters(
    engine.StaticParameter{Value: []string{"git", "curl"}},
    engine.StaticParameter{Value: ""}, // package manager: "" = auto-detect, "apt", or "brew"
)
```

## Utilities

### WaitAction

Waits for specified duration.

```go
action, err := utility.NewWaitAction(logger).WithParameters(
    engine.StaticParameter{Value: time.Second * 5},
)
```

### PrerequisiteCheckAction

Conditional execution based on custom checks. The check function returns `(abortTask bool, err error)`.

```go
action, err := utility.NewPrerequisiteCheckAction(logger).WithParameters(
    engine.StaticParameter{Value: "checking disk space"},
    engine.StaticParameter{Value: utility.PrerequisiteCheckFunc(func(ctx context.Context, logger *slog.Logger) (bool, error) {
        // Return (true, nil) to abort the task, (false, nil) to continue
        return false, nil
    })},
)
```

### FetchInterfacesAction

Gets network interface information.

```go
utility.NewFetchInterfacesAction(logger)
```

### ReadMACAddressAction

Reads MAC address of network interface.

```go
action, err := utility.NewReadMacAction(logger).WithParameters(
    engine.StaticParameter{Value: "eth0"},
)
```

## Examples

See `tasks/` directory for complete workflows:

- `NewFileOperationsTask()` - File management pipeline
- `NewDockerSetupTask()` - Docker environment setup
- `NewPackageManagementTask()` - Package installation
- `NewExtractOperationsTask()` - Archive extraction
