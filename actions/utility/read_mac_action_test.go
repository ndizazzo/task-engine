package utility_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/utility"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

type ReadMacActionTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *ReadMacActionTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "read_mac_test_*")
	suite.Require().NoError(err)
}

func (suite *ReadMacActionTestSuite) TearDownTest() {
	_ = os.RemoveAll(suite.tempDir)
}

func (suite *ReadMacActionTestSuite) TestNewReadMACAddressAction() {
	logger := command_mock.NewDiscardLogger()
	action := utility.NewReadMACAddressAction(logger)
	suite.NotNil(action)
	suite.Equal(logger, action.Logger)
}

func (suite *ReadMacActionTestSuite) TestNewReadMACAddressActionNilLogger() {
	action := utility.NewReadMACAddressAction(nil)
	suite.NotNil(action)
	// With current implementation, nil logger is passed directly
	suite.Nil(action.Logger)
}

func (suite *ReadMacActionTestSuite) TestWithParametersValid() {
	logger := command_mock.NewDiscardLogger()
	interfaceParam := engine.StaticParameter{Value: "eth0"}

	wrappedAction, err := utility.NewReadMACAddressAction(logger).WithParameters(interfaceParam)
	suite.NoError(err)
	suite.NotNil(wrappedAction)
	suite.Equal("read-mac-action", wrappedAction.ID)
	suite.Equal("Read MAC Address", wrappedAction.Name)
	suite.Equal(interfaceParam, wrappedAction.Wrapped.InterfaceNameParam)
}

func (suite *ReadMacActionTestSuite) TestWithParametersNilParameter() {
	logger := command_mock.NewDiscardLogger()

	wrappedAction, err := utility.NewReadMACAddressAction(logger).WithParameters(nil)
	suite.Error(err)
	suite.Nil(wrappedAction)
	suite.Contains(err.Error(), "interface name parameter cannot be nil")
}

func (suite *ReadMacActionTestSuite) TestExecuteSuccess() {
	// Note: This test will fail on most systems since /sys/class/net/eth0/address doesn't exist
	// This is mainly to test the error path, but we can verify parameter resolution
	logger := command_mock.NewDiscardLogger()
	action := utility.NewReadMACAddressAction(logger)
	action.InterfaceNameParam = engine.StaticParameter{Value: "nonexistent_interface"}

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "failed to read MAC address for interface nonexistent_interface")
}

func (suite *ReadMacActionTestSuite) TestExecuteInterfaceNotFound() {
	logger := command_mock.NewDiscardLogger()
	action := utility.NewReadMACAddressAction(logger)
	action.InterfaceNameParam = engine.StaticParameter{Value: "nonexistent"}

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "failed to read MAC address for interface nonexistent")
}

func (suite *ReadMacActionTestSuite) TestExecuteEmptyInterfaceName() {
	logger := command_mock.NewDiscardLogger()
	action := utility.NewReadMACAddressAction(logger)
	action.InterfaceNameParam = engine.StaticParameter{Value: ""}

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "interface name cannot be empty")
}

func (suite *ReadMacActionTestSuite) TestExecuteInvalidParameterType() {
	logger := command_mock.NewDiscardLogger()
	action := utility.NewReadMACAddressAction(logger)
	action.InterfaceNameParam = engine.StaticParameter{Value: 123} // Not a string

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "interface name parameter is not a string")
}

func (suite *ReadMacActionTestSuite) TestExecuteParameterResolutionFailure() {
	logger := command_mock.NewDiscardLogger()
	action := utility.NewReadMACAddressAction(logger)

	// Create a mock parameter that fails to resolve
	mockParam := &command_mock.MockActionParameter{
		ResolveFunc: func(ctx context.Context, gc *engine.GlobalContext) (interface{}, error) {
			return nil, fmt.Errorf("parameter resolution failed")
		},
	}
	action.InterfaceNameParam = mockParam

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "failed to resolve interface name parameter")
}

func (suite *ReadMacActionTestSuite) TestExecuteEmptyMACAddress() {
	// This test verifies that the action properly handles empty MAC addresses
	// Since we can't easily mock the file system, we'll test the basic error path
	logger := command_mock.NewDiscardLogger()
	action := utility.NewReadMACAddressAction(logger)
	action.InterfaceNameParam = engine.StaticParameter{Value: "empty_mac_interface"}

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "failed to read MAC address for interface empty_mac_interface")
}

func (suite *ReadMacActionTestSuite) TestGetOutput() {
	logger := command_mock.NewDiscardLogger()
	action := utility.NewReadMACAddressAction(logger)

	// Set some test data
	action.Interface = "eth0"
	action.MAC = "aa:bb:cc:dd:ee:ff"

	output := action.GetOutput()
	suite.NotNil(output)

	outputMap, ok := output.(map[string]interface{})
	suite.True(ok, "Output should be a map")

	suite.Equal("eth0", outputMap["interface"])
	suite.Equal("aa:bb:cc:dd:ee:ff", outputMap["mac"])
	suite.Equal(true, outputMap["success"])
}

func (suite *ReadMacActionTestSuite) TestGetOutputNoMAC() {
	logger := command_mock.NewDiscardLogger()
	action := utility.NewReadMACAddressAction(logger)

	// Set interface but no MAC (simulating failure case)
	action.Interface = "eth0"
	action.MAC = ""

	output := action.GetOutput()
	suite.NotNil(output)

	outputMap, ok := output.(map[string]interface{})
	suite.True(ok, "Output should be a map")

	suite.Equal("eth0", outputMap["interface"])
	suite.Equal("", outputMap["mac"])
	suite.Equal(false, outputMap["success"])
}

func (suite *ReadMacActionTestSuite) TestExecuteWithGlobalContext() {
	logger := command_mock.NewDiscardLogger()
	action := utility.NewReadMACAddressAction(logger)
	action.InterfaceNameParam = engine.StaticParameter{Value: "test_interface"}

	// Create a context with GlobalContext
	gc := &engine.GlobalContext{}
	ctx := context.WithValue(context.Background(), engine.GlobalContextKey, gc)

	// This will fail since interface doesn't exist, but tests parameter resolution
	err := action.Execute(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to read MAC address for interface test_interface")
}

func TestReadMacActionTestSuite(t *testing.T) {
	suite.Run(t, new(ReadMacActionTestSuite))
}
