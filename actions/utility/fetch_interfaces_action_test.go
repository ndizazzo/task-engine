package utility_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ndizazzo/task-engine/actions/utility"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// FetchInterfacesActionTestSuite tests the FetchNetworkInterfacesAction functionality
type FetchInterfacesActionTestSuite struct {
	suite.Suite
}

// TestFetchInterfacesActionTestSuite runs the FetchInterfacesAction test suite
func TestFetchInterfacesActionTestSuite(t *testing.T) {
	suite.Run(t, new(FetchInterfacesActionTestSuite))
}

func (suite *FetchInterfacesActionTestSuite) TestFetchNetworkInterfacesAction() {
	tempDir := suite.T().TempDir()

	// Mock network interfaces as directories
	mockInterfaces := []string{"enp1s0", "enx001a2b3c4d", "wlan0", "docker0", "lo"}
	for _, iface := range mockInterfaces {
		err := os.Mkdir(filepath.Join(tempDir, iface), 0755)
		require.NoError(suite.T(), err)
	}

	// Create wireless directory for wlan0 to mark it as wireless
	err := os.Mkdir(filepath.Join(tempDir, "wlan0", "wireless"), 0755)
	require.NoError(suite.T(), err)

	action := utility.NewFetchNetInterfacesAction(tempDir, nil)

	err = action.Wrapped.Execute(context.Background())
	assert.NoError(suite.T(), err)

	expected := []string{"enp1s0", "enx001a2b3c4d", "wlan0", "docker0", "lo"}
	assert.Equal(suite.T(), expected, action.Wrapped.Interfaces)
}
