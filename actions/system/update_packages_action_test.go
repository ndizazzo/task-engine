package system

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

// UpdatePackagesActionTestSuite tests the UpdatePackagesAction
type UpdatePackagesActionTestSuite struct {
	suite.Suite
}

// TestUpdatePackagesActionTestSuite runs the UpdatePackagesAction test suite
func TestUpdatePackagesActionTestSuite(t *testing.T) {
	suite.Run(t, new(UpdatePackagesActionTestSuite))
}

// Tests for new constructor pattern with parameters
func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_WithParameters() {
	logger := slog.Default()

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl", "wget"}}, // packageNames
		task_engine.StaticParameter{Value: "apt"},                    // packageManager
	)

	suite.Require().NoError(err)
	suite.NotNil(action)
	suite.Equal("update-packages-action", action.ID)
	suite.NotNil(action.Wrapped)
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_WithNilLogger() {
	constructor := NewUpdatePackagesAction(nil)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl"}},
		task_engine.StaticParameter{Value: "apt"},
	)

	suite.Require().NoError(err)
	suite.NotNil(action)
	suite.NotNil(action.Wrapped.Logger)
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_WithAptManager() {
	logger := mocks.NewDiscardLogger()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "update").Return("Reading package lists... Done", nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "install", "-y", "curl", "wget").Return("Packages installed successfully", nil)

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl", "wget"}}, // packageNames
		task_engine.StaticParameter{Value: "apt"},                    // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal([]string{"curl", "wget"}, action.Wrapped.PackageNames)
	suite.Equal(AptPackageManager, action.Wrapped.PackageManager)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_WithBrewManager() {
	logger := mocks.NewDiscardLogger()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "brew", "install", "curl", "wget").Return("Packages installed successfully", nil)

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl", "wget"}}, // packageNames
		task_engine.StaticParameter{Value: "brew"},                   // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal([]string{"curl", "wget"}, action.Wrapped.PackageNames)
	suite.Equal(BrewPackageManager, action.Wrapped.PackageManager)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_WithPackageNamesAsString() {
	logger := mocks.NewDiscardLogger()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "update").Return("Reading package lists... Done", nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "install", "-y", "curl", "wget").Return("Packages installed successfully", nil)

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: "curl,wget"}, // packageNames as comma-separated string
		task_engine.StaticParameter{Value: "apt"},       // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal([]string{"curl", "wget"}, action.Wrapped.PackageNames)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_WithPackageNamesAsSpaceSeparated() {
	logger := mocks.NewDiscardLogger()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "brew", "install", "curl", "wget").Return("Packages installed successfully", nil)

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: "curl wget"}, // packageNames as space-separated string
		task_engine.StaticParameter{Value: "brew"},      // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal([]string{"curl", "wget"}, action.Wrapped.PackageNames)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_EmptyPackageNames() {
	logger := mocks.NewDiscardLogger()

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{}}, // empty packageNames
		task_engine.StaticParameter{Value: "apt"},      // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "no package names provided")
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_UnsupportedPackageManager() {
	logger := mocks.NewDiscardLogger()

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl"}}, // packageNames
		task_engine.StaticParameter{Value: "unsupported"},    // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "unsupported package manager")
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_EmptyPackageManager() {
	logger := mocks.NewDiscardLogger()

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl"}}, // packageNames
		task_engine.StaticParameter{Value: ""},               // empty packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "unsupported operating system for package management")
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_AptUpdateFailure() {
	logger := mocks.NewDiscardLogger()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "update").Return("", errors.New("update failed"))

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl"}}, // packageNames
		task_engine.StaticParameter{Value: "apt"},            // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to update apt package list")

	mockRunner.AssertExpectations(suite.T())
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_AptInstallFailure() {
	logger := mocks.NewDiscardLogger()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "update").Return("Reading package lists... Done", nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "install", "-y", "curl").Return("", errors.New("install failed"))

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl"}}, // packageNames
		task_engine.StaticParameter{Value: "apt"},            // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to install packages with apt")

	mockRunner.AssertExpectations(suite.T())
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_BrewInstallFailure() {
	logger := mocks.NewDiscardLogger()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "brew", "install", "curl").Return("", errors.New("install failed"))

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl"}}, // packageNames
		task_engine.StaticParameter{Value: "brew"},           // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to install packages with brew")

	mockRunner.AssertExpectations(suite.T())
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_InvalidParameterTypes() {
	logger := mocks.NewDiscardLogger()

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: 123}, // packageNames should be []string or string, not int
		task_engine.StaticParameter{Value: "apt"},
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "package names parameter is not a string slice or string")
	action, err = constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl"}},
		task_engine.StaticParameter{Value: 123}, // packageManager should be string, not int
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "package manager parameter resolved to non-string value")
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_NilParameters() {
	logger := mocks.NewDiscardLogger()

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		nil, // nil packageNames parameter
		nil, // nil packageManager parameter
	)

	suite.Require().NoError(err)

	// Set up the action to have valid values directly (since parameters are nil)
	action.Wrapped.PackageNames = []string{"curl"}
	action.Wrapped.PackageManager = AptPackageManager

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "update").Return("Reading package lists... Done", nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "install", "-y", "curl").Return("Packages installed successfully", nil)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *UpdatePackagesActionTestSuite) TestUpdatePackagesAction_GetOutput() {
	action := &UpdatePackagesAction{
		PackageNames:   []string{"curl", "wget"},
		PackageManager: AptPackageManager,
	}

	output := action.GetOutput()

	suite.IsType(map[string]interface{}{}, output)
	outputMap := output.(map[string]interface{})

	suite.Equal([]string{"curl", "wget"}, outputMap["packages"])
	suite.Equal("apt", outputMap["packageManager"])
	suite.Equal(true, outputMap["success"])
}

func (suite *UpdatePackagesActionTestSuite) TestDetectPackageManager() {
	// This test will vary based on the OS the test is run on
	packageManager := detectPackageManager()

	// On any system, the detected package manager should be one of the supported ones or empty
	suite.True(packageManager == AptPackageManager || packageManager == BrewPackageManager || packageManager == "")
}

func (suite *UpdatePackagesActionTestSuite) TestIsCommandAvailable() {
	available := isCommandAvailable("ls")
	suite.True(available)
	available = isCommandAvailable("nonexistentcommand12345")
	suite.False(available)
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_WithMultiplePackages() {
	logger := mocks.NewDiscardLogger()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "update").Return("Reading package lists... Done", nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "apt", "install", "-y", "curl", "wget", "git").Return("Packages installed successfully", nil)

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl", "wget", "git"}}, // multiple packages
		task_engine.StaticParameter{Value: "apt"},                           // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal([]string{"curl", "wget", "git"}, action.Wrapped.PackageNames)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *UpdatePackagesActionTestSuite) TestNewUpdatePackagesActionConstructor_Execute_WithContext() {
	logger := mocks.NewDiscardLogger()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "brew", "install", "curl").Return("Package installed successfully", nil)

	constructor := NewUpdatePackagesAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"curl"}}, // packageNames
		task_engine.StaticParameter{Value: "brew"},           // packageManager
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	ctx := context.Background()
	err = action.Wrapped.Execute(ctx)

	suite.NoError(err)

	mockRunner.AssertExpectations(suite.T())
}
