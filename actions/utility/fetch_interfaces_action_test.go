package utility_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/utility"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// FetchInterfacesActionTestSuite tests the FetchNetworkInterfacesAction functionality
type FetchInterfacesActionTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func (suite *FetchInterfacesActionTestSuite) SetupTest() {
	suite.logger = command_mock.NewDiscardLogger()
}

// TestFetchInterfacesActionTestSuite runs the FetchInterfacesAction test suite
func TestFetchInterfacesActionTestSuite(t *testing.T) {
	suite.Run(t, new(FetchInterfacesActionTestSuite))
}

func (suite *FetchInterfacesActionTestSuite) TestFetchNetworkInterfacesAction_WithPresetInterfaces() {
	mockInterfaces := []string{"enp1s0", "enx001a2b3c4d", "wlan0", "docker0", "lo"}
	action, err := utility.NewFetchNetInterfacesAction(suite.logger).WithParameters(
		task_engine.StaticParameter{Value: "/sys/class/net"}, // device path (won't be used)
		task_engine.StaticParameter{Value: mockInterfaces},   // preset interfaces
	)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())
	assert.NoError(suite.T(), execErr)

	expected := []string{"enp1s0", "enx001a2b3c4d", "wlan0", "docker0", "lo"}
	assert.Equal(suite.T(), expected, action.Wrapped.Interfaces)
}

func (suite *FetchInterfacesActionTestSuite) TestFetchNetworkInterfacesAction_WithScanning() {
	tempDir := suite.T().TempDir()

	// Mock network interfaces as directories
	mockInterfaces := []string{"enp1s0", "enx001a2b3c4d", "wlan0", "docker0", "lo"}
	for _, iface := range mockInterfaces {
		err := os.Mkdir(filepath.Join(tempDir, iface), 0o755)
		require.NoError(suite.T(), err)
	}

	// Create wireless directory for wlan0 to mark it as wireless
	err := os.Mkdir(filepath.Join(tempDir, "wlan0", "wireless"), 0o755)
	require.NoError(suite.T(), err)

	action, err := utility.NewFetchNetInterfacesAction(suite.logger).WithParameters(
		task_engine.StaticParameter{Value: tempDir}, // device path to scan
		nil, // no preset interfaces - will scan device path
	)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())
	assert.NoError(suite.T(), execErr)
	assert.Contains(suite.T(), action.Wrapped.Interfaces, "enp1s0")
	assert.Contains(suite.T(), action.Wrapped.Interfaces, "wlan0")
	assert.Equal(suite.T(), tempDir, action.Wrapped.NetDevicePath)
}

func (suite *FetchInterfacesActionTestSuite) TestFetchNetInterfacesAction_GetOutput() {
	action := &utility.FetchNetInterfacesAction{
		Interfaces: []string{"eth0", "lo"},
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal(2, m["count"])
	suite.Equal(true, m["success"])
	suite.Len(m["interfaces"], 2)
}
