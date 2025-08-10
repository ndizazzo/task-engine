package system_test

import (
	"context"
	"testing"

	"github.com/ndizazzo/task-engine/actions/system"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceStatusActionTestSuite struct {
	suite.Suite
	mockProcessor *command_mock.MockCommandRunner
}

func (suite *ServiceStatusActionTestSuite) SetupTest() {
	suite.mockProcessor = new(command_mock.MockCommandRunner)
}

func (suite *ServiceStatusActionTestSuite) TestGetSingleServiceStatus() {
	serviceName := "lemony-agent.service"
	expectedOutput := `LoadState=loaded
ActiveState=inactive
SubState=dead
Description=Lemony Update Agent
FragmentPath=/etc/systemd/system/lemony-agent.service
Vendor=disabled; vendor preset: enabled`

	logger := command_mock.NewDiscardLogger()
	action := system.NewGetServiceStatusAction(logger, serviceName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", serviceName).Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result
	suite.Len(action.Wrapped.ServiceStatuses, 1)
	service := action.Wrapped.ServiceStatuses[0]
	suite.Equal(serviceName, service.Name)
	suite.Equal("Lemony Update Agent", service.Description)
	suite.Equal("loaded", service.Loaded)
	suite.Equal("inactive (dead)", service.Active)
	suite.Equal("dead", service.Sub)
	suite.Equal("/etc/systemd/system/lemony-agent.service", service.Path)
	suite.Equal("disabled; vendor preset: enabled", service.Vendor)
	suite.True(service.Exists)
}

func (suite *ServiceStatusActionTestSuite) TestGetMultipleServiceStatuses() {
	serviceNames := []string{"lemony-agent.service", "networkd.service"}

	// Mock responses for each service
	lemonyOutput := `LoadState=loaded
ActiveState=inactive
SubState=dead
Description=Lemony Update Agent
FragmentPath=/etc/systemd/system/lemony-agent.service
Vendor=disabled; vendor preset: enabled`

	networkdOutput := `Unit networkd.service could not be found.`

	logger := command_mock.NewDiscardLogger()
	action := system.NewGetServiceStatusAction(logger, serviceNames...)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", "lemony-agent.service").Return(lemonyOutput, nil)
	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", "networkd.service").Return(networkdOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed results
	suite.Len(action.Wrapped.ServiceStatuses, 2)

	// Check lemony-agent service
	lemonyService := action.Wrapped.ServiceStatuses[0]
	suite.Equal("lemony-agent.service", lemonyService.Name)
	suite.Equal("Lemony Update Agent", lemonyService.Description)
	suite.Equal("loaded", lemonyService.Loaded)
	suite.Equal("inactive (dead)", lemonyService.Active)
	suite.True(lemonyService.Exists)

	// Check networkd service (doesn't exist)
	networkdService := action.Wrapped.ServiceStatuses[1]
	suite.Equal("networkd.service", networkdService.Name)
	suite.False(networkdService.Exists)
}

func (suite *ServiceStatusActionTestSuite) TestGetAllServicesStatusNotSupported() {
	logger := command_mock.NewDiscardLogger()
	action := system.NewGetAllServicesStatusAction(logger)

	err := action.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "getting all services status is not supported")
}

func (suite *ServiceStatusActionTestSuite) TestServiceWithDifferentStates() {
	serviceName := "sshd.service"
	expectedOutput := `LoadState=loaded
ActiveState=active
SubState=running
Description=OpenBSD Secure Shell server
FragmentPath=/lib/systemd/system/sshd.service
Vendor=enabled; vendor preset: enabled`

	logger := command_mock.NewDiscardLogger()
	action := system.NewGetServiceStatusAction(logger, serviceName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", serviceName).Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result
	suite.Len(action.Wrapped.ServiceStatuses, 1)
	service := action.Wrapped.ServiceStatuses[0]
	suite.Equal(serviceName, service.Name)
	suite.Equal("OpenBSD Secure Shell server", service.Description)
	suite.Equal("loaded", service.Loaded)
	suite.Equal("active (running)", service.Active)
	suite.Equal("running", service.Sub)
	suite.Equal("/lib/systemd/system/sshd.service", service.Path)
	suite.Equal("enabled; vendor preset: enabled", service.Vendor)
	suite.True(service.Exists)
}

func (suite *ServiceStatusActionTestSuite) TestServiceWithMinimalOutput() {
	serviceName := "minimal.service"
	expectedOutput := `LoadState=loaded
ActiveState=inactive
SubState=dead
Description=Minimal Service
FragmentPath=/etc/systemd/system/minimal.service`

	logger := command_mock.NewDiscardLogger()
	action := system.NewGetServiceStatusAction(logger, serviceName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", serviceName).Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result
	suite.Len(action.Wrapped.ServiceStatuses, 1)
	service := action.Wrapped.ServiceStatuses[0]
	suite.Equal(serviceName, service.Name)
	suite.Equal("Minimal Service", service.Description)
	suite.Equal("loaded", service.Loaded)
	suite.Equal("inactive (dead)", service.Active)
	suite.Equal("dead", service.Sub)
	suite.Equal("/etc/systemd/system/minimal.service", service.Path)
	suite.Equal("", service.Vendor)
	suite.True(service.Exists)
}

func (suite *ServiceStatusActionTestSuite) TestServiceWithComplexVendorInfo() {
	serviceName := "complex.service"
	expectedOutput := `LoadState=loaded
ActiveState=active
SubState=running
Description=Complex Service with Vendor Info
FragmentPath=/lib/systemd/system/complex.service
Vendor=Custom Vendor; vendor preset: disabled; custom setting: enabled`

	logger := command_mock.NewDiscardLogger()
	action := system.NewGetServiceStatusAction(logger, serviceName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", serviceName).Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result
	suite.Len(action.Wrapped.ServiceStatuses, 1)
	service := action.Wrapped.ServiceStatuses[0]
	suite.Equal(serviceName, service.Name)
	suite.Equal("Complex Service with Vendor Info", service.Description)
	suite.Equal("loaded", service.Loaded)
	suite.Equal("active (running)", service.Active)
	suite.Equal("running", service.Sub)
	suite.Equal("/lib/systemd/system/complex.service", service.Path)
	suite.Equal("Custom Vendor; vendor preset: disabled; custom setting: enabled", service.Vendor)
	suite.True(service.Exists)
}

func (suite *ServiceStatusActionTestSuite) TestNonExistentService() {
	serviceName := "nonexistent.service"
	expectedOutput := `Unit nonexistent.service could not be found.`

	logger := command_mock.NewDiscardLogger()
	action := system.NewGetServiceStatusAction(logger, serviceName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", serviceName).Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result
	suite.Len(action.Wrapped.ServiceStatuses, 1)
	service := action.Wrapped.ServiceStatuses[0]
	suite.Equal(serviceName, service.Name)
	suite.False(service.Exists)
}

func (suite *ServiceStatusActionTestSuite) TestCommandFailure() {
	serviceName := "failing.service"

	logger := command_mock.NewDiscardLogger()
	action := system.NewGetServiceStatusAction(logger, serviceName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", serviceName).Return("", assert.AnError)

	err := action.Execute(context.Background())

	suite.NoError(err) // Should not fail, should return service with Exists=false
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result
	suite.Len(action.Wrapped.ServiceStatuses, 1)
	service := action.Wrapped.ServiceStatuses[0]
	suite.Equal(serviceName, service.Name)
	suite.False(service.Exists)
}

func (suite *ServiceStatusActionTestSuite) TestContextCancellation() {
	serviceName := "test.service"

	logger := command_mock.NewDiscardLogger()
	action := system.NewGetServiceStatusAction(logger, serviceName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	suite.mockProcessor.On("RunCommandWithContext", ctx, "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", serviceName).Return("", context.Canceled)

	err := action.Execute(ctx)

	suite.NoError(err) // Should not fail, should return service with Exists=false
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result
	suite.Len(action.Wrapped.ServiceStatuses, 1)
	service := action.Wrapped.ServiceStatuses[0]
	suite.Equal(serviceName, service.Name)
	suite.False(service.Exists)
}

func (suite *ServiceStatusActionTestSuite) TestEmptyServiceName() {
	serviceName := ""

	logger := command_mock.NewDiscardLogger()
	action := system.NewGetServiceStatusAction(logger, serviceName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", serviceName).Return("", assert.AnError)

	err := action.Execute(context.Background())

	suite.NoError(err) // Should not fail, should return service with Exists=false
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result
	suite.Len(action.Wrapped.ServiceStatuses, 1)
	service := action.Wrapped.ServiceStatuses[0]
	suite.Equal(serviceName, service.Name)
	suite.False(service.Exists)
}

func (suite *ServiceStatusActionTestSuite) TestServiceWithUnicodeCharacters() {
	serviceName := "unicode-服务.service"
	expectedOutput := `LoadState=loaded
ActiveState=active
SubState=running
Description=Unicode Service 服务
FragmentPath=/lib/systemd/system/unicode-服务.service
Vendor=enabled; vendor preset: enabled`

	logger := command_mock.NewDiscardLogger()
	action := system.NewGetServiceStatusAction(logger, serviceName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "systemctl", "show", "--property=LoadState,ActiveState,SubState,Description,FragmentPath,Vendor,UnitFileState", serviceName).Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result
	suite.Len(action.Wrapped.ServiceStatuses, 1)
	service := action.Wrapped.ServiceStatuses[0]
	suite.Equal(serviceName, service.Name)
	suite.Equal("Unicode Service 服务", service.Description)
	suite.Equal("loaded", service.Loaded)
	suite.Equal("active (running)", service.Active)
	suite.Equal("running", service.Sub)
	suite.Equal("/lib/systemd/system/unicode-服务.service", service.Path)
	suite.Equal("enabled; vendor preset: enabled", service.Vendor)
	suite.True(service.Exists)
}

func TestServiceStatusActionTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceStatusActionTestSuite))
}
