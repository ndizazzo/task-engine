# Built-in Actions

Complete list of available actions. See `tasks/` directory for usage examples.

## File Operations

### CreateDirectoriesAction

Creates multiple directories.

```go
file.NewCreateDirectoriesAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/root"},
    engine.StaticParameter{Value: []string{"src", "tests", "docs"}},
)
```

### WriteFileAction

Writes content to files.

```go
file.NewWriteFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/file"},
    engine.StaticParameter{Value: []byte("content")},
    true, // overwrite
    nil,  // inputBuffer
)
```

### ReadFileAction

Reads file contents.

```go
var content []byte
file.NewReadFileAction("/path/to/file", &content, logger)
```

### CompressFileAction

Compresses files (gzip).

```go
file.NewCompressFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/source/file"},
    engine.StaticParameter{Value: "/dest/file.gz"},
    file.GzipCompression,
)
```

### DecompressFileAction

Decompresses files with auto-detection.

```go
file.NewDecompressFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/source/file.gz"},
    engine.StaticParameter{Value: "/dest/file"},
    "", // auto-detect
)
```

### CopyFileAction

Copies files and directories.

```go
file.NewCopyFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/source"},
    engine.StaticParameter{Value: "/dest"},
    true,  // createDirs
    false, // recursive
)
```

### DeletePathAction

Safely deletes files and directories.

```go
file.NewDeletePathAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/delete"},
    true,  // recursive
    false, // dryRun
)
```

### ReplaceLinesAction

Replaces text using regex patterns.

```go
file.NewReplaceLinesAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/file"},
    map[*regexp.Regexp]engine.ActionParameter{
        regexp.MustCompile("old"): engine.StaticParameter{Value: "new"},
    },
)
```

### ChangeOwnershipAction

Changes file ownership.

```go
file.NewChangeOwnershipAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path"},
    engine.StaticParameter{Value: "user"},
    engine.StaticParameter{Value: "group"},
    true, // recursive
)
```

### ChangePermissionsAction

Changes file permissions.

```go
file.NewChangePermissionsAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path"},
    engine.StaticParameter{Value: "755"},
    true, // recursive
)
```

### MoveFileAction

Moves/renames files.

```go
file.NewMoveFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/source"},
    engine.StaticParameter{Value: "/dest"},
    true, // createDirs
)
```

### ExtractFileAction

Extracts archives (tar, zip) with security features.

```go
file.NewExtractFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/archive.tar"},
    engine.StaticParameter{Value: "/extract/dir"},
    file.AutoDetect, // or file.TarArchive, file.ZipArchive
)
```

## Docker Operations

### GetContainerStateAction

Gets container status by ID or name.

```go
docker.NewGetContainerStateAction(logger, "container1", "container2")
```

### GetAllContainersStateAction

Gets status of all containers.

```go
docker.NewGetAllContainersStateAction(logger)
```

### DockerComposeUpAction

Starts Docker Compose services.

```go
docker.NewDockerComposeUpAction(logger).WithParameters(
    engine.StaticParameter{Value: []string{"web", "db"}},
    engine.StaticParameter{Value: "/path/to/compose"},
)
```

### DockerComposeDownAction

Stops Docker Compose services.

```go
docker.NewDockerComposeDownAction(logger).WithParameters(
    engine.StaticParameter{Value: []string{"web"}},
    engine.StaticParameter{Value: "/path/to/compose"},
)
```

### DockerComposeExecAction

Executes commands in containers.

```go
docker.NewDockerComposeExecAction(logger).WithParameters(
    engine.StaticParameter{Value: "web"},
    engine.StaticParameter{Value: []string{"ls", "-la"}},
    engine.StaticParameter{Value: "/path/to/compose"},
)
```

### DockerRunAction

Runs Docker containers.

```go
docker.NewDockerRunAction(logger).WithParameters(
    engine.StaticParameter{Value: "nginx:latest"},
    engine.StaticParameter{Value: []string{"nginx", "-g", "daemon off;"}},
    engine.StaticParameter{Value: []string{"-p", "8080:80"},
    nil, // inputBuffer
)
```

### CheckContainerHealthAction

Performs health checks with retries.

```go
docker.NewCheckContainerHealthAction(logger).WithParameters(
    engine.StaticParameter{Value: "container"},
    engine.StaticParameter{Value: 3},           // maxRetries
    engine.StaticParameter{Value: time.Second}, // retryDelay
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

### GetServiceStatusAction

Gets systemd service status.

```go
system.NewGetServiceStatusAction(logger, "nginx", "mysql")
```

### ManageServiceAction

Controls systemd services.

```go
system.NewManageServiceAction(logger).WithParameters(
    engine.StaticParameter{Value: "nginx"},
    engine.StaticParameter{Value: "restart"},
)
```

### ShutdownAction

Shuts down or restarts system.

```go
system.NewShutdownAction(logger).WithParameters(
    engine.StaticParameter{Value: "restart"},
    engine.StaticParameter{Value: "5"}, // delay in minutes
)
```

### UpdatePackagesAction

Installs packages (apt/brew).

```go
system.NewUpdatePackagesAction(logger, []string{"git", "curl"})
```

## Utilities

### WaitAction

Waits for specified duration.

```go
utility.NewWaitAction(logger).WithParameters(
    engine.StaticParameter{Value: time.Second * 5},
)
```

### PrerequisiteCheckAction

Conditional execution based on custom checks.

```go
utility.NewPrerequisiteCheckAction(logger).WithParameters(
    engine.StaticParameter{Value: func(ctx context.Context) (bool, error) {
        // custom check logic
        return true, nil
    }},
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
utility.NewReadMacAction(logger).WithParameters(
    engine.StaticParameter{Value: "eth0"},
)
```

## Examples

See `tasks/` directory for complete workflows:

- `NewFileOperationsTask()` - File management pipeline
- `NewDockerSetupTask()` - Docker environment setup
- `NewPackageManagementTask()` - Package installation
- `NewExtractOperationsTask()` - Archive extraction
