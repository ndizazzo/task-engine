package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerPsAction(t *testing.T) {
	logger := slog.Default()

	action := NewDockerPsAction(logger)

	assert.NotNil(t, action)
	assert.Equal(t, "docker-ps-action", action.ID)
	assert.False(t, action.Wrapped.All)
	assert.Empty(t, action.Wrapped.Filter)
	assert.Empty(t, action.Wrapped.Format)
	assert.Equal(t, 0, action.Wrapped.Last)
	assert.False(t, action.Wrapped.Latest)
	assert.False(t, action.Wrapped.NoTrunc)
	assert.False(t, action.Wrapped.Quiet)
	assert.False(t, action.Wrapped.Size)
}

func TestNewDockerPsActionWithOptions(t *testing.T) {
	logger := slog.Default()

	action := NewDockerPsAction(logger,
		WithPsAll(),
		WithPsFilter("status=running"),
		WithPsFormat("table {{.Names}}\t{{.Status}}"),
		WithPsLast(5),
		WithPsLatest(),
		WithPsNoTrunc(),
		WithPsQuiet(),
		WithPsSize(),
	)

	assert.NotNil(t, action)
	assert.True(t, action.Wrapped.All)
	assert.Equal(t, "status=running", action.Wrapped.Filter)
	assert.Equal(t, "table {{.Names}}\t{{.Status}}", action.Wrapped.Format)
	assert.Equal(t, 5, action.Wrapped.Last)
	assert.True(t, action.Wrapped.Latest)
	assert.True(t, action.Wrapped.NoTrunc)
	assert.True(t, action.Wrapped.Quiet)
	assert.True(t, action.Wrapped.Size)
}

func TestDockerPsAction_Execute_Success(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1
def456ghi789   redis     "docker-entrypoint.s"    1 hour ago      Up 1 hour      6379/tcp   myapp_redis_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return(expectedOutput, nil)

	action := NewDockerPsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Containers, 2)

	// Check first container
	assert.Equal(t, "abc123def456", action.Wrapped.Containers[0].ContainerID)
	assert.Equal(t, "nginx", action.Wrapped.Containers[0].Image)
	assert.Equal(t, "nginx -g 'daemon off", action.Wrapped.Containers[0].Command)
	assert.Equal(t, "2 hours ago", action.Wrapped.Containers[0].Created)
	assert.Equal(t, "Up 2 hours", action.Wrapped.Containers[0].Status)
	assert.Equal(t, "0.0.0.0:8080->80/tcp", action.Wrapped.Containers[0].Ports)
	assert.Equal(t, "myapp_web_1", action.Wrapped.Containers[0].Names)

	// Check second container
	assert.Equal(t, "def456ghi789", action.Wrapped.Containers[1].ContainerID)
	assert.Equal(t, "redis", action.Wrapped.Containers[1].Image)
	assert.Equal(t, "docker-entrypoint.s", action.Wrapped.Containers[1].Command)
	assert.Equal(t, "1 hour ago", action.Wrapped.Containers[1].Created)
	assert.Equal(t, "Up 1 hour", action.Wrapped.Containers[1].Status)
	assert.Equal(t, "6379/tcp", action.Wrapped.Containers[1].Ports)
	assert.Equal(t, "myapp_redis_1", action.Wrapped.Containers[1].Names)

	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_WithAll(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS                     PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours                0.0.0.0:8080->80/tcp   myapp_web_1
def456ghi789   redis     "docker-entrypoint.s"    1 hour ago      Exited (0) 1 hour ago     6379/tcp   myapp_redis_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--all").Return(expectedOutput, nil)

	action := NewDockerPsAction(logger, WithPsAll())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Containers, 2)
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_WithFilter(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--filter", "status=running").Return(expectedOutput, nil)

	action := NewDockerPsAction(logger, WithPsFilter("status=running"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Containers, 1)
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_WithFormat(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `myapp_web_1:Up 2 hours
myapp_redis_1:Up 1 hour`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--format", "{{.Names}}:{{.Status}}").Return(expectedOutput, nil)

	action := NewDockerPsAction(logger, WithPsFormat("{{.Names}}:{{.Status}}"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	// With custom format, we don't parse the output into structured data
	assert.Empty(t, action.Wrapped.Containers)
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_WithLast(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--last", "3").Return(expectedOutput, nil)

	action := NewDockerPsAction(logger, WithPsLast(3))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Containers, 1)
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_WithLatest(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--latest").Return(expectedOutput, nil)

	action := NewDockerPsAction(logger, WithPsLatest())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Containers, 1)
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_WithNoTrunc(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID                                                                    IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def4567890123456789012345678901234567890123456789012345678901234567890   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--no-trunc").Return(expectedOutput, nil)

	action := NewDockerPsAction(logger, WithPsNoTrunc())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Containers, 1)
	assert.Equal(t, "abc123def4567890123456789012345678901234567890123456789012345678901234567890", action.Wrapped.Containers[0].ContainerID)
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_WithQuiet(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `abc123def456
def456ghi789`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--quiet").Return(expectedOutput, nil)

	action := NewDockerPsAction(logger, WithPsQuiet())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	// With quiet mode, we don't parse the output into structured data
	assert.Empty(t, action.Wrapped.Containers)
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_WithSize(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES     SIZE
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1   1.23MB (virtual 133MB)`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--size").Return(expectedOutput, nil)

	action := NewDockerPsAction(logger, WithPsSize())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Containers, 1)
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_CommandError(t *testing.T) {
	logger := slog.Default()
	expectedError := "permission denied"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return("", errors.New(expectedError))

	action := NewDockerPsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
	assert.Empty(t, action.Wrapped.Containers)
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_ContextCancellation(t *testing.T) {
	logger := slog.Default()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return("", context.Canceled)

	action := NewDockerPsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_parseContainers(t *testing.T) {
	tests := []struct {
		name               string
		output             string
		expectedContainers []Container
	}{
		{
			name:               "empty output",
			output:             "",
			expectedContainers: []Container(nil),
		},
		{
			name:               "only header",
			output:             "CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES",
			expectedContainers: []Container(nil),
		},
		{
			name: "single container",
			output: `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1`,
			expectedContainers: []Container{
				{
					ContainerID: "abc123def456",
					Image:       "nginx",
					Command:     "nginx -g 'daemon off",
					Created:     "2 hours ago",
					Status:      "Up 2 hours",
					Ports:       "0.0.0.0:8080->80/tcp",
					Names:       "myapp_web_1",
				},
			},
		},
		{
			name: "multiple containers",
			output: `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1
def456ghi789   redis     "docker-entrypoint.s"    1 hour ago      Up 1 hour      6379/tcp   myapp_redis_1`,
			expectedContainers: []Container{
				{
					ContainerID: "abc123def456",
					Image:       "nginx",
					Command:     "nginx -g 'daemon off",
					Created:     "2 hours ago",
					Status:      "Up 2 hours",
					Ports:       "0.0.0.0:8080->80/tcp",
					Names:       "myapp_web_1",
				},
				{
					ContainerID: "def456ghi789",
					Image:       "redis",
					Command:     "docker-entrypoint.s",
					Created:     "1 hour ago",
					Status:      "Up 1 hour",
					Ports:       "6379/tcp",
					Names:       "myapp_redis_1",
				},
			},
		},
		{
			name: "container without ports",
			output: `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     myapp_web_1`,
			expectedContainers: []Container{
				{
					ContainerID: "abc123def456",
					Image:       "nginx",
					Command:     "nginx -g 'daemon off",
					Created:     "2 hours ago",
					Status:      "Up 2 hours",
					Ports:       "",
					Names:       "myapp_web_1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			action := NewDockerPsAction(logger)

			action.Wrapped.parseContainers(tt.output)

			assert.Equal(t, tt.expectedContainers, action.Wrapped.Containers)
		})
	}
}

func TestDockerPsAction_parseContainerLine(t *testing.T) {
	tests := []struct {
		name              string
		line              string
		expectedContainer *Container
	}{
		{
			name: "valid container line",
			line: "abc123def456   nginx     \"nginx -g 'daemon off\"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1",
			expectedContainer: &Container{
				ContainerID: "abc123def456",
				Image:       "nginx",
				Command:     "nginx -g 'daemon off",
				Created:     "2 hours ago",
				Status:      "Up 2 hours",
				Ports:       "0.0.0.0:8080->80/tcp",
				Names:       "myapp_web_1",
			},
		},
		{
			name: "container without ports",
			line: "abc123def456   nginx     \"nginx -g 'daemon off\"   2 hours ago     Up 2 hours     myapp_web_1",
			expectedContainer: &Container{
				ContainerID: "abc123def456",
				Image:       "nginx",
				Command:     "nginx -g 'daemon off",
				Created:     "2 hours ago",
				Status:      "Up 2 hours",
				Ports:       "",
				Names:       "myapp_web_1",
			},
		},
		{
			name: "container with complex command",
			line: "abc123def456   node     \"node /app/server.js\"   30 minutes ago   Up 30 minutes   3000/tcp   myapp_node_1",
			expectedContainer: &Container{
				ContainerID: "abc123def456",
				Image:       "node",
				Command:     "node /app/server.js",
				Created:     "30 minutes ago",
				Status:      "Up 30 minutes",
				Ports:       "3000/tcp",
				Names:       "myapp_node_1",
			},
		},
		{
			name: "container with multiple names",
			line: "abc123def456   nginx     \"nginx -g 'daemon off\"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1,myapp_web",
			expectedContainer: &Container{
				ContainerID: "abc123def456",
				Image:       "nginx",
				Command:     "nginx -g 'daemon off",
				Created:     "2 hours ago",
				Status:      "Up 2 hours",
				Ports:       "0.0.0.0:8080->80/tcp",
				Names:       "myapp_web_1,myapp_web",
			},
		},
		{
			name:              "insufficient parts",
			line:              "abc123def456 nginx",
			expectedContainer: nil,
		},
		{
			name:              "empty line",
			line:              "",
			expectedContainer: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			action := NewDockerPsAction(logger)

			result := action.Wrapped.parseContainerLine(tt.line)

			assert.Equal(t, tt.expectedContainer, result)
		})
	}
}

func TestDockerPsAction_Execute_EmptyOutput(t *testing.T) {
	logger := slog.Default()
	expectedOutput := ""

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return(expectedOutput, nil)

	action := NewDockerPsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Empty(t, action.Wrapped.Containers)
	mockRunner.AssertExpectations(t)
}

func TestDockerPsAction_Execute_OutputWithTrailingWhitespace(t *testing.T) {
	logger := slog.Default()
	rawOutput := "CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES\nabc123def456   nginx     \"nginx -g 'daemon off\"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1\n  \n  "

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return(rawOutput, nil)

	action := NewDockerPsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, rawOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Containers, 1)
	assert.Equal(t, "abc123def456", action.Wrapped.Containers[0].ContainerID)
	mockRunner.AssertExpectations(t)
}
