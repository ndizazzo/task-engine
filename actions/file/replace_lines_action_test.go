package file_test

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"testing"

	engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ReplaceLinesTestSuite defines the test suite for ReplaceLinesAction.
type ReplaceLinesTestSuite struct {
	suite.Suite
}

func writeTestFile(t *testing.T, content string) string {
	tempFile, err := os.CreateTemp("", "testfile*.conf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_, err = tempFile.WriteString(content)
	require.NoError(t, err)
	tempFile.Close()
	return tempFile.Name()
}

func readTestFile(t *testing.T, path string) string {
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var content string
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	return content
}

// TestReplaceLines tests various scenarios for ReplaceLinesAction.
func (suite *ReplaceLinesTestSuite) TestReplaceLines() {
	tests := []struct {
		name            string
		inputContent    string
		patterns        map[*regexp.Regexp]string
		expectedContent string
	}{
		{
			name:         "Replace matching line",
			inputContent: "interface=wlan0\ndhcp-option=3\n",
			patterns: map[*regexp.Regexp]string{
				regexp.MustCompile(`^interface=.*$`): "interface=eth0",
			},
			expectedContent: "interface=eth0\ndhcp-option=3\n",
		},
		{
			name:         "No matches",
			inputContent: "dhcp-option=3\nlease-time=86400\n",
			patterns: map[*regexp.Regexp]string{
				regexp.MustCompile(`^interface=.*$`): "interface=eth0",
			},
			expectedContent: "dhcp-option=3\nlease-time=86400\n",
		},
		{
			name:         "Multiple matches",
			inputContent: "interface=wlan0\ninterface=eth1\n",
			patterns: map[*regexp.Regexp]string{
				regexp.MustCompile(`^interface=.*$`): "interface=eth0",
			},
			expectedContent: "interface=eth0\ninterface=eth0\n",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			filePath := writeTestFile(suite.T(), tt.inputContent)
			defer os.Remove(filePath)

			action := &engine.Action[*file.ReplaceLinesAction]{
				ID: "test-replace-lines",
				Wrapped: &file.ReplaceLinesAction{
					FilePath:        filePath,
					ReplacePatterns: tt.patterns,
				},
			}

			err := action.Execute(context.Background())
			suite.NoError(err)

			output := readTestFile(suite.T(), filePath)
			suite.Equal(tt.expectedContent, output)
		})
	}
}

// TestReplaceLinesTestSuite runs the ReplaceLinesTestSuite.
func TestReplaceLinesTestSuite(t *testing.T) {
	suite.Run(t, new(ReplaceLinesTestSuite))
}
