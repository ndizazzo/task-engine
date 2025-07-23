package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerComposeLsAction(t *testing.T) {
	logger := slog.Default()

	action := NewDockerComposeLsAction(logger)

	assert.NotNil(t, action)
	assert.Equal(t, "docker-compose-ls-action", action.ID)
	assert.False(t, action.Wrapped.All)
	assert.Empty(t, action.Wrapped.Filter)
	assert.Empty(t, action.Wrapped.Format)
	assert.False(t, action.Wrapped.Quiet)
	assert.Empty(t, action.Wrapped.WorkingDir)
}

func TestNewDockerComposeLsActionWithOptions(t *testing.T) {
	logger := slog.Default()

	action := NewDockerComposeLsAction(logger,
		WithComposeAll(),
		WithComposeFilter("name=myapp"),
		WithComposeFormat("table {{.Name}}\t{{.Status}}"),
		WithComposeQuiet(),
		WithWorkingDir("/path/to/compose"),
	)

	assert.NotNil(t, action)
	assert.True(t, action.Wrapped.All)
	assert.Equal(t, "name=myapp", action.Wrapped.Filter)
	assert.Equal(t, "table {{.Name}}\t{{.Status}}", action.Wrapped.Format)
	assert.True(t, action.Wrapped.Quiet)
	assert.Equal(t, "/path/to/compose", action.Wrapped.WorkingDir)
}

func TestDockerComposeLsAction_Execute_Success(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml
testapp             stopped             /path/to/compose.yml,/path/to/override.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Stacks, 2)

	// Check first stack
	assert.Equal(t, "myapp", action.Wrapped.Stacks[0].Name)
	assert.Equal(t, "running", action.Wrapped.Stacks[0].Status)
	assert.Equal(t, "/path/to/docker-compose.yml", action.Wrapped.Stacks[0].ConfigFiles)

	// Check second stack
	assert.Equal(t, "testapp", action.Wrapped.Stacks[1].Name)
	assert.Equal(t, "stopped", action.Wrapped.Stacks[1].Status)
	assert.Equal(t, "/path/to/compose.yml,/path/to/override.yml", action.Wrapped.Stacks[1].ConfigFiles)

	mockRunner.AssertExpectations(t)
}

func TestDockerComposeLsAction_Execute_WithAll(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml
testapp             stopped             /path/to/compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls", "--all").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger, WithComposeAll())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Stacks, 2)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposeLsAction_Execute_WithFilter(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls", "--filter", "name=myapp").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger, WithComposeFilter("name=myapp"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Stacks, 1)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposeLsAction_Execute_WithFormat(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `myapp:running
testapp:stopped`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls", "--format", "{{.Name}}:{{.Status}}").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger, WithComposeFormat("{{.Name}}:{{.Status}}"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	// With custom format, we don't parse the output into structured data
	assert.Empty(t, action.Wrapped.Stacks)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposeLsAction_Execute_WithQuiet(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `myapp
testapp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls", "--quiet").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger, WithComposeQuiet())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	// With quiet mode, we don't parse the output into structured data
	assert.Empty(t, action.Wrapped.Stacks)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposeLsAction_Execute_CommandError(t *testing.T) {
	logger := slog.Default()
	expectedError := "permission denied"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return("", errors.New(expectedError))

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
	assert.Empty(t, action.Wrapped.Stacks)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposeLsAction_Execute_ContextCancellation(t *testing.T) {
	logger := slog.Default()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return("", context.Canceled)

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	mockRunner.AssertExpectations(t)
}

func TestDockerComposeLsAction_parseStacks(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedStacks []ComposeStack
	}{
		{
			name:           "empty output",
			output:         "",
			expectedStacks: []ComposeStack(nil),
		},
		{
			name:           "only header",
			output:         "NAME                STATUS              CONFIG FILES",
			expectedStacks: []ComposeStack(nil),
		},
		{
			name: "single stack",
			output: `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml`,
			expectedStacks: []ComposeStack{
				{
					Name:        "myapp",
					Status:      "running",
					ConfigFiles: "/path/to/docker-compose.yml",
				},
			},
		},
		{
			name: "multiple stacks",
			output: `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml
testapp             stopped             /path/to/compose.yml,/path/to/override.yml`,
			expectedStacks: []ComposeStack{
				{
					Name:        "myapp",
					Status:      "running",
					ConfigFiles: "/path/to/docker-compose.yml",
				},
				{
					Name:        "testapp",
					Status:      "stopped",
					ConfigFiles: "/path/to/compose.yml,/path/to/override.yml",
				},
			},
		},
		{
			name: "stack with multiple config files",
			output: `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/compose.yml,/path/to/override.yml,/path/to/prod.yml`,
			expectedStacks: []ComposeStack{
				{
					Name:        "myapp",
					Status:      "running",
					ConfigFiles: "/path/to/compose.yml,/path/to/override.yml,/path/to/prod.yml",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			action := NewDockerComposeLsAction(logger)

			action.Wrapped.parseStacks(tt.output)

			assert.Equal(t, tt.expectedStacks, action.Wrapped.Stacks)
		})
	}
}

func TestDockerComposeLsAction_parseStackLine(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		expectedStack *ComposeStack
	}{
		{
			name: "valid stack line",
			line: "myapp               running             /path/to/docker-compose.yml",
			expectedStack: &ComposeStack{
				Name:        "myapp",
				Status:      "running",
				ConfigFiles: "/path/to/docker-compose.yml",
			},
		},
		{
			name: "stack with multiple config files",
			line: "testapp             stopped             /path/to/compose.yml,/path/to/override.yml",
			expectedStack: &ComposeStack{
				Name:        "testapp",
				Status:      "stopped",
				ConfigFiles: "/path/to/compose.yml,/path/to/override.yml",
			},
		},
		{
			name: "stack with spaces in config files",
			line: "myapp               running             /path/to/my compose.yml",
			expectedStack: &ComposeStack{
				Name:        "myapp",
				Status:      "running",
				ConfigFiles: "/path/to/my compose.yml",
			},
		},
		{
			name:          "insufficient parts",
			line:          "myapp running",
			expectedStack: nil,
		},
		{
			name:          "empty line",
			line:          "",
			expectedStack: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			action := NewDockerComposeLsAction(logger)

			result := action.Wrapped.parseStackLine(tt.line)

			assert.Equal(t, tt.expectedStack, result)
		})
	}
}

func TestDockerComposeLsAction_Execute_EmptyOutput(t *testing.T) {
	logger := slog.Default()
	expectedOutput := ""

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Empty(t, action.Wrapped.Stacks)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposeLsAction_Execute_OutputWithTrailingWhitespace(t *testing.T) {
	logger := slog.Default()
	rawOutput := "NAME                STATUS              CONFIG FILES\nmyapp               running             /path/to/docker-compose.yml\n  \n  "

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(rawOutput, nil)

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, rawOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Stacks, 1)
	assert.Equal(t, "myapp", action.Wrapped.Stacks[0].Name)
	mockRunner.AssertExpectations(t)
}
