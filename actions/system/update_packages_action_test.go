package system

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/suite"

	task_engine "github.com/ndizazzo/task-engine"
	command_mock "github.com/ndizazzo/task-engine/mocks"
)

// MockCommandRunner is a mock implementation of CommandRunner for testing
type MockCommandRunner struct {
	commands []string
	args     [][]string
	outputs  []string
	errors   []error
	index    int
}

func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		commands: make([]string, 0),
		args:     make([][]string, 0),
		outputs:  make([]string, 0),
		errors:   make([]error, 0),
		index:    0,
	}
}

func (m *MockCommandRunner) AddCommand(command string, args []string, output string, err error) {
	m.commands = append(m.commands, command)
	m.args = append(m.args, args)
	m.outputs = append(m.outputs, output)
	m.errors = append(m.errors, err)
}

func (m *MockCommandRunner) RunCommand(command string, args ...string) (string, error) {
	return m.RunCommandWithContext(context.Background(), command, args...)
}

func (m *MockCommandRunner) RunCommandWithContext(ctx context.Context, command string, args ...string) (string, error) {
	if m.index >= len(m.commands) {
		return "", errors.New("unexpected command call")
	}

	expectedCommand := m.commands[m.index]
	expectedArgs := m.args[m.index]
	output := m.outputs[m.index]
	err := m.errors[m.index]

	if command != expectedCommand {
		return "", errors.New("unexpected command")
	}

	if len(args) != len(expectedArgs) {
		return "", errors.New("unexpected args length")
	}

	for i, arg := range args {
		if i < len(expectedArgs) && arg != expectedArgs[i] {
			return "", errors.New("unexpected arg")
		}
	}

	m.index++
	return output, err
}

func (m *MockCommandRunner) RunCommandInDir(workingDir string, command string, args ...string) (string, error) {
	return m.RunCommandInDirWithContext(context.Background(), workingDir, command, args...)
}

func (m *MockCommandRunner) RunCommandInDirWithContext(ctx context.Context, workingDir string, command string, args ...string) (string, error) {
	return m.RunCommandWithContext(ctx, command, args...)
}

type UpdatePackagesTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func (suite *UpdatePackagesTestSuite) SetupTest() {
	suite.logger = command_mock.NewDiscardLogger()
}

func (suite *UpdatePackagesTestSuite) TestNewUpdatePackagesActionValidParameters() {
	packageNames := []string{"package1", "package2"}
	action := NewUpdatePackagesAction(packageNames, suite.logger)

	suite.NotNil(action)
	suite.NotNil(action.Wrapped)
	suite.Equal(packageNames, action.Wrapped.PackageNames)
	suite.NotEmpty(action.Wrapped.PackageManager)
	suite.NotNil(action.Wrapped.CommandRunner)
}

func (suite *UpdatePackagesTestSuite) TestNewUpdatePackagesActionNilLogger() {
	packageNames := []string{"package1"}
	action := NewUpdatePackagesAction(packageNames, nil)

	suite.NotNil(action)
	suite.NotNil(action.Wrapped)
	suite.NotNil(action.Wrapped.Logger)
}

func (suite *UpdatePackagesTestSuite) TestNewUpdatePackagesActionEmptyPackageList() {
	packageNames := []string{}
	action := NewUpdatePackagesAction(packageNames, suite.logger)

	suite.NotNil(action)
	suite.NotNil(action.Wrapped)
	suite.Empty(action.Wrapped.PackageNames)
}

func (suite *UpdatePackagesTestSuite) TestExecuteEmptyPackageList() {
	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{},
		PackageManager: AptPackageManager,
		CommandRunner:  NewMockCommandRunner(),
	}

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "no package names provided")
}

func (suite *UpdatePackagesTestSuite) TestExecuteUnsupportedPackageManager() {
	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"package1"},
		PackageManager: "",
		CommandRunner:  NewMockCommandRunner(),
	}

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "unsupported operating system for package management")
}

func (suite *UpdatePackagesTestSuite) TestExecuteUnsupportedPackageManagerType() {
	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"package1"},
		PackageManager: "unsupported",
		CommandRunner:  NewMockCommandRunner(),
	}

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "unsupported package manager")
}

func (suite *UpdatePackagesTestSuite) TestExecuteAptSuccess() {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddCommand("apt", []string{"update"}, "Package list updated", nil)
	mockRunner.AddCommand("apt", []string{"install", "-y", "package1", "package2"}, "Packages installed", nil)

	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"package1", "package2"},
		PackageManager: AptPackageManager,
		CommandRunner:  mockRunner,
	}

	err := action.Execute(context.Background())
	suite.NoError(err)
}

func (suite *UpdatePackagesTestSuite) TestExecuteAptUpdateFailure() {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddCommand("apt", []string{"update"}, "", errors.New("update failed"))

	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"package1"},
		PackageManager: AptPackageManager,
		CommandRunner:  mockRunner,
	}

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to update apt package list")
}

func (suite *UpdatePackagesTestSuite) TestExecuteAptInstallFailure() {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddCommand("apt", []string{"update"}, "Package list updated", nil)
	mockRunner.AddCommand("apt", []string{"install", "-y", "package1"}, "", errors.New("install failed"))

	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"package1"},
		PackageManager: AptPackageManager,
		CommandRunner:  mockRunner,
	}

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to install packages with apt")
}

func (suite *UpdatePackagesTestSuite) TestExecuteBrewSuccess() {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddCommand("brew", []string{"install", "package1", "package2"}, "Packages installed", nil)

	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"package1", "package2"},
		PackageManager: BrewPackageManager,
		CommandRunner:  mockRunner,
	}

	err := action.Execute(context.Background())
	suite.NoError(err)
}

func (suite *UpdatePackagesTestSuite) TestExecuteBrewFailure() {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddCommand("brew", []string{"install", "package1"}, "", errors.New("install failed"))

	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"package1"},
		PackageManager: BrewPackageManager,
		CommandRunner:  mockRunner,
	}

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to install packages with brew")
}

func (suite *UpdatePackagesTestSuite) TestDetectPackageManagerLinux() {
	// This test will only work on Linux systems
	if runtime.GOOS != "linux" {
		suite.T().Skip("Skipping Linux-specific test on non-Linux system")
	}

	// Test that detectPackageManager works on Linux
	// Note: This is a basic test that doesn't mock the command availability
	packageManager := detectPackageManager()

	// On Linux, it should either be apt or empty (if apt is not available)
	if packageManager != "" {
		suite.Equal(AptPackageManager, packageManager)
	}
}

func (suite *UpdatePackagesTestSuite) TestDetectPackageManagerDarwin() {
	// This test will only work on macOS systems
	if runtime.GOOS != "darwin" {
		suite.T().Skip("Skipping macOS-specific test on non-macOS system")
	}

	// Test that detectPackageManager works on macOS
	packageManager := detectPackageManager()

	// On macOS, it should either be brew or empty (if brew is not available)
	if packageManager != "" {
		suite.Equal(BrewPackageManager, packageManager)
	}
}

func (suite *UpdatePackagesTestSuite) TestIsCommandAvailable() {
	// Test with a command that should always be available
	available := isCommandAvailable("ls")
	suite.True(available)

	// Test with a command that should not be available
	available = isCommandAvailable("nonexistentcommand12345")
	suite.False(available)
}

func (suite *UpdatePackagesTestSuite) TestExecuteWithContext() {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddCommand("apt", []string{"update"}, "Package list updated", nil)
	mockRunner.AddCommand("apt", []string{"install", "-y", "package1"}, "Package installed", nil)

	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"package1"},
		PackageManager: AptPackageManager,
		CommandRunner:  mockRunner,
	}

	ctx := context.Background()
	err := action.Execute(ctx)
	suite.NoError(err)
}

func (suite *UpdatePackagesTestSuite) TestExecuteAptWithMultiplePackages() {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddCommand("apt", []string{"update"}, "Package list updated", nil)
	mockRunner.AddCommand("apt", []string{"install", "-y", "package1", "package2", "package3"}, "Packages installed", nil)

	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"package1", "package2", "package3"},
		PackageManager: AptPackageManager,
		CommandRunner:  mockRunner,
	}

	err := action.Execute(context.Background())
	suite.NoError(err)
}

func (suite *UpdatePackagesTestSuite) TestExecuteBrewWithMultiplePackages() {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddCommand("brew", []string{"install", "package1", "package2", "package3"}, "Packages installed", nil)

	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"package1", "package2", "package3"},
		PackageManager: BrewPackageManager,
		CommandRunner:  mockRunner,
	}

	err := action.Execute(context.Background())
	suite.NoError(err)
}

func (suite *UpdatePackagesTestSuite) TestExecuteAptWithOutput() {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddCommand("apt", []string{"update"}, "Reading package lists... Done", nil)
	mockRunner.AddCommand("apt", []string{"install", "-y", "curl"}, "curl is already the newest version", nil)

	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"curl"},
		PackageManager: AptPackageManager,
		CommandRunner:  mockRunner,
	}

	err := action.Execute(context.Background())
	suite.NoError(err)
}

func (suite *UpdatePackagesTestSuite) TestExecuteBrewWithOutput() {
	mockRunner := NewMockCommandRunner()
	mockRunner.AddCommand("brew", []string{"install", "wget"}, "wget is already installed", nil)

	action := &UpdatePackagesAction{
		BaseAction:     task_engine.BaseAction{Logger: suite.logger},
		PackageNames:   []string{"wget"},
		PackageManager: BrewPackageManager,
		CommandRunner:  mockRunner,
	}

	err := action.Execute(context.Background())
	suite.NoError(err)
}

func TestUpdatePackagesTestSuite(t *testing.T) {
	suite.Run(t, new(UpdatePackagesTestSuite))
}
