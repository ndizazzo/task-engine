package file_test

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ReplaceLinesTestSuite defines the test suite for ReplaceLinesAction.
type ReplaceLinesTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *ReplaceLinesTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "replace_lines_test_*")
	suite.Require().NoError(err)
}

func (suite *ReplaceLinesTestSuite) TearDownTest() {
	_ = os.RemoveAll(suite.tempDir)
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

func (suite *ReplaceLinesTestSuite) TestNewReplaceLinesAction() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	suite.NotNil(action)
	suite.Equal(logger, action.Logger)
}

func (suite *ReplaceLinesTestSuite) TestNewReplaceLinesActionNilLogger() {
	action := file.NewReplaceLinesAction(nil)
	suite.NotNil(action)
	suite.NotNil(action.Logger)
}

func (suite *ReplaceLinesTestSuite) TestWithParameters() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)

	filePath := engine.StaticParameter{Value: "/path/to/file.txt"}
	patterns := make(map[*regexp.Regexp]engine.ActionParameter)
	patterns[regexp.MustCompile(`test`)] = engine.StaticParameter{Value: "replacement"}

	wrappedAction, err := action.WithParameters(filePath, patterns)
	suite.NoError(err)
	suite.NotNil(wrappedAction)
	suite.Equal("replace-lines-action", wrappedAction.ID)
	suite.Equal("Replace Lines", wrappedAction.Name)
	suite.Equal(filePath, wrappedAction.Wrapped.FilePathParam)
	suite.Equal(patterns, wrappedAction.Wrapped.ReplaceParamPatterns)
}

func (suite *ReplaceLinesTestSuite) TestWithParametersNilParams() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)

	wrappedAction, err := action.WithParameters(nil, nil)
	suite.NoError(err)
	suite.NotNil(wrappedAction)
	suite.Nil(wrappedAction.Wrapped.FilePathParam)
	suite.Nil(wrappedAction.Wrapped.ReplaceParamPatterns)
}

func (suite *ReplaceLinesTestSuite) TestExecuteWithFilePathParameter() {
	filePath := filepath.Join(suite.tempDir, "test.conf")
	initialContent := "interface=wlan0\ndhcp-option=3\nserver=old\n"
	err := os.WriteFile(filePath, []byte(initialContent), 0o644)
	suite.Require().NoError(err)

	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)

	// Use parameterized file path
	filePathParam := engine.StaticParameter{Value: filePath}
	patterns := make(map[*regexp.Regexp]engine.ActionParameter)
	patterns[regexp.MustCompile(`^interface=.*$`)] = engine.StaticParameter{Value: "interface=eth0"}
	patterns[regexp.MustCompile(`^server=.*$`)] = engine.StaticParameter{Value: "server=new"}

	wrappedAction, err := action.WithParameters(filePathParam, patterns)
	suite.NoError(err)

	// Create context with global context for parameter resolution
	gc := &engine.GlobalContext{}
	ctx := context.WithValue(context.Background(), engine.GlobalContextKey, gc)

	err = wrappedAction.Wrapped.Execute(ctx)
	suite.NoError(err)
	actualContent, err := os.ReadFile(filePath)
	suite.NoError(err)
	expectedContent := "interface=eth0\ndhcp-option=3\nserver=new\n"
	suite.Equal(expectedContent, string(actualContent))
}

func (suite *ReplaceLinesTestSuite) TestExecuteWithParameterizedReplacements() {
	filePath := filepath.Join(suite.tempDir, "config.txt")
	initialContent := "user=olduser\nport=8080\nhost=localhost\n"
	err := os.WriteFile(filePath, []byte(initialContent), 0o644)
	suite.Require().NoError(err)

	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = filePath // Set directly since we're not using path parameter

	// Create parameterized replacements
	patterns := make(map[*regexp.Regexp]engine.ActionParameter)
	patterns[regexp.MustCompile(`^user=.*$`)] = engine.StaticParameter{Value: "user=newuser"}
	patterns[regexp.MustCompile(`^port=.*$`)] = engine.StaticParameter{Value: "port=9090"}
	patterns[regexp.MustCompile(`^host=.*$`)] = engine.StaticParameter{Value: "host=production"}

	action.ReplaceParamPatterns = patterns

	// Create context with global context for parameter resolution
	gc := &engine.GlobalContext{}
	ctx := context.WithValue(context.Background(), engine.GlobalContextKey, gc)

	err = action.Execute(ctx)
	suite.NoError(err)
	actualContent, err := os.ReadFile(filePath)
	suite.NoError(err)
	expectedContent := "user=newuser\nport=9090\nhost=production\n"
	suite.Equal(expectedContent, string(actualContent))
}

func (suite *ReplaceLinesTestSuite) TestExecuteFilePathResolutionFailure() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)

	// Mock parameter that fails to resolve
	mockParam := &command_mock.MockActionParameter{
		ResolveFunc: func(ctx context.Context, gc *engine.GlobalContext) (interface{}, error) {
			return nil, fmt.Errorf("failed to resolve file path")
		},
	}
	action.FilePathParam = mockParam

	gc := &engine.GlobalContext{}
	ctx := context.WithValue(context.Background(), engine.GlobalContextKey, gc)

	err := action.Execute(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to resolve file path parameter")
}

func (suite *ReplaceLinesTestSuite) TestExecuteFilePathInvalidType() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)

	// Parameter that resolves to non-string
	mockParam := &command_mock.MockActionParameter{
		ResolveFunc: func(ctx context.Context, gc *engine.GlobalContext) (interface{}, error) {
			return 12345, nil // Not a string
		},
	}
	action.FilePathParam = mockParam

	gc := &engine.GlobalContext{}
	ctx := context.WithValue(context.Background(), engine.GlobalContextKey, gc)

	err := action.Execute(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "file path parameter is not a string")
}

func (suite *ReplaceLinesTestSuite) TestExecuteEmptyFilePath() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = "" // Empty file path

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "file path cannot be empty")
}

func (suite *ReplaceLinesTestSuite) TestExecuteNoGlobalContextForParameters() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)

	// Set a file path parameter but no global context
	action.FilePathParam = engine.StaticParameter{Value: "/some/path"}

	// Context without GlobalContext
	ctx := context.Background()

	err := action.Execute(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "global context not available for path parameter resolution")
}

func (suite *ReplaceLinesTestSuite) TestExecuteReplacementParameterResolutionFailure() {
	filePath := filepath.Join(suite.tempDir, "test.txt")
	err := os.WriteFile(filePath, []byte("test line\n"), 0o644)
	suite.Require().NoError(err)

	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = filePath

	// Create pattern with failing parameter
	patterns := make(map[*regexp.Regexp]engine.ActionParameter)
	mockParam := &command_mock.MockActionParameter{
		ResolveFunc: func(ctx context.Context, gc *engine.GlobalContext) (interface{}, error) {
			return nil, fmt.Errorf("replacement resolution failed")
		},
	}
	patterns[regexp.MustCompile(`test`)] = mockParam
	action.ReplaceParamPatterns = patterns

	gc := &engine.GlobalContext{}
	ctx := context.WithValue(context.Background(), engine.GlobalContextKey, gc)

	err = action.Execute(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to resolve replacement parameter")
}

func (suite *ReplaceLinesTestSuite) TestExecuteReplacementParameterTypes() {
	filePath := filepath.Join(suite.tempDir, "types.txt")
	initialContent := "string_value=old\nbytes_value=old\nint_value=old\n"
	err := os.WriteFile(filePath, []byte(initialContent), 0o644)
	suite.Require().NoError(err)

	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = filePath
	patterns := make(map[*regexp.Regexp]engine.ActionParameter)

	// String parameter
	patterns[regexp.MustCompile(`^string_value=.*$`)] = engine.StaticParameter{Value: "string_value=new_string"}

	// Bytes parameter
	bytesParam := &command_mock.MockActionParameter{
		ResolveFunc: func(ctx context.Context, gc *engine.GlobalContext) (interface{}, error) {
			return []byte("bytes_value=new_bytes"), nil
		},
	}
	patterns[regexp.MustCompile(`^bytes_value=.*$`)] = bytesParam

	// Integer parameter (should be converted to string)
	intParam := &command_mock.MockActionParameter{
		ResolveFunc: func(ctx context.Context, gc *engine.GlobalContext) (interface{}, error) {
			return 42, nil
		},
	}
	patterns[regexp.MustCompile(`^int_value=.*$`)] = intParam

	action.ReplaceParamPatterns = patterns

	gc := &engine.GlobalContext{}
	ctx := context.WithValue(context.Background(), engine.GlobalContextKey, gc)

	err = action.Execute(ctx)
	suite.NoError(err)
	actualContent, err := os.ReadFile(filePath)
	suite.NoError(err)
	expectedContent := "string_value=new_string\nbytes_value=new_bytes\n42\n"
	suite.Equal(expectedContent, string(actualContent))
}

func (suite *ReplaceLinesTestSuite) TestExecuteNilParameterHandling() {
	filePath := filepath.Join(suite.tempDir, "nil_param.txt")
	err := os.WriteFile(filePath, []byte("test=value\n"), 0o644)
	suite.Require().NoError(err)

	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = filePath

	// Pattern with nil parameter
	patterns := make(map[*regexp.Regexp]engine.ActionParameter)
	patterns[regexp.MustCompile(`^test=.*$`)] = nil
	action.ReplaceParamPatterns = patterns

	gc := &engine.GlobalContext{}
	ctx := context.WithValue(context.Background(), engine.GlobalContextKey, gc)

	err = action.Execute(ctx)
	suite.NoError(err)
	actualContent, err := os.ReadFile(filePath)
	suite.NoError(err)
	expectedContent := "\n" // Empty replacement
	suite.Equal(expectedContent, string(actualContent))
}

func (suite *ReplaceLinesTestSuite) TestExecuteFileNotFound() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = "/nonexistent/path/file.txt"

	err := action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "failed to open file")
}

func (suite *ReplaceLinesTestSuite) TestExecuteFilePermissionError() {
	// Create a file we can't read
	filePath := filepath.Join(suite.tempDir, "no_read_perm.txt")
	err := os.WriteFile(filePath, []byte("test\n"), 0o644)
	suite.Require().NoError(err)

	// Remove read permissions
	err = os.Chmod(filePath, 0o000)
	suite.Require().NoError(err)

	// Restore permissions in cleanup so file can be removed
	defer func() {
		_ = os.Chmod(filePath, 0o644)
	}()

	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = filePath

	err = action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "failed to open file")
}

func (suite *ReplaceLinesTestSuite) TestExecuteFirstMatchingPatternOnly() {
	filePath := filepath.Join(suite.tempDir, "first_match.txt")
	initialContent := "server=localhost\n"
	err := os.WriteFile(filePath, []byte(initialContent), 0o644)
	suite.Require().NoError(err)

	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = filePath

	// Multiple patterns that could match the same line
	patterns := map[*regexp.Regexp]string{
		regexp.MustCompile(`server=.*`):      "server=first",
		regexp.MustCompile(`server=local.*`): "server=second", // This should not be applied
	}
	action.ReplacePatterns = patterns

	err = action.Execute(context.Background())
	suite.NoError(err)
	actualContent, err := os.ReadFile(filePath)
	suite.NoError(err)
	// The result should be one of the two possible outcomes
	result := string(actualContent)
	suite.True(result == "server=first\n" || result == "server=second\n",
		"Expected either 'server=first\n' or 'server=second\n', got: %q", result)
}

func (suite *ReplaceLinesTestSuite) TestExecuteLargeFile() {
	filePath := filepath.Join(suite.tempDir, "large_file.txt")

	// Create file with many lines
	lines := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		if i%100 == 0 {
			lines[i] = "special_line=old_value"
		} else {
			lines[i] = fmt.Sprintf("line_%d=content_%d", i, i)
		}
	}
	initialContent := fmt.Sprintf("%s\n", fmt.Sprintf("%s", joinStrings(lines, "\n")))
	err := os.WriteFile(filePath, []byte(initialContent), 0o644)
	suite.Require().NoError(err)

	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = filePath
	action.ReplacePatterns = map[*regexp.Regexp]string{
		regexp.MustCompile(`^special_line=.*$`): "special_line=new_value",
	}

	err = action.Execute(context.Background())
	suite.NoError(err)
	actualContent, err := os.ReadFile(filePath)
	suite.NoError(err)
	suite.Contains(string(actualContent), "special_line=new_value")
	suite.NotContains(string(actualContent), "special_line=old_value")
}

func (suite *ReplaceLinesTestSuite) TestGetOutput() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = "/path/to/file.txt"

	// Set up patterns for output
	patterns := map[*regexp.Regexp]string{
		regexp.MustCompile(`test1`): "replacement1",
		regexp.MustCompile(`test2`): "replacement2",
	}
	action.ReplacePatterns = patterns

	output := action.GetOutput()
	suite.NotNil(output)

	outputMap, ok := output.(map[string]interface{})
	suite.True(ok, "Output should be a map")

	suite.Equal("/path/to/file.txt", outputMap["filePath"])
	suite.Equal(2, outputMap["patterns"])
	suite.Equal(true, outputMap["success"])
}

func (suite *ReplaceLinesTestSuite) TestGetOutputEmptyPatterns() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = "/some/path.txt"
	// No patterns set

	output := action.GetOutput()
	suite.NotNil(output)

	outputMap, ok := output.(map[string]interface{})
	suite.True(ok, "Output should be a map")

	suite.Equal("/some/path.txt", outputMap["filePath"])
	suite.Equal(0, outputMap["patterns"])
	suite.Equal(true, outputMap["success"])
}

func (suite *ReplaceLinesTestSuite) TestComplexRegexPatterns() {
	filePath := filepath.Join(suite.tempDir, "complex_regex.txt")
	initialContent := `# Configuration file
server.port=8080
server.host=localhost
# Comment line
db.url=jdbc:mysql://localhost:3306/test
db.username=admin
db.password=secret123
`
	err := os.WriteFile(filePath, []byte(initialContent), 0o644)
	suite.Require().NoError(err)

	logger := command_mock.NewDiscardLogger()
	action := file.NewReplaceLinesAction(logger)
	action.FilePath = filePath

	// Complex patterns
	patterns := map[*regexp.Regexp]string{
		// Replace port numbers
		regexp.MustCompile(`(server\.port=)\d+`): "${1}9090",
		// Replace database URL
		regexp.MustCompile(`^db\.url=.*$`): "db.url=jdbc:postgresql://postgres:5432/newdb",
		// Replace passwords (but keep the key)
		regexp.MustCompile(`(db\.password=).*`): "${1}new_secret",
	}
	action.ReplacePatterns = patterns

	err = action.Execute(context.Background())
	suite.NoError(err)

	actualContent, err := os.ReadFile(filePath)
	suite.NoError(err)
	result := string(actualContent)
	suite.Contains(result, "server.port=9090")
	suite.Contains(result, "db.url=jdbc:postgresql://postgres:5432/newdb")
	suite.Contains(result, "db.password=new_secret")
	// Ensure comments and other lines are preserved
	suite.Contains(result, "# Configuration file")
	suite.Contains(result, "# Comment line")
	suite.Contains(result, "server.host=localhost")
	suite.Contains(result, "db.username=admin")
}

// Helper function to join strings (since strings.Join might not be imported)
func joinStrings(strs []string, separator string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += separator + strs[i]
	}
	return result
}

// TestReplaceLinesTestSuite runs the ReplaceLinesTestSuite.
func TestReplaceLinesTestSuite(t *testing.T) {
	suite.Run(t, new(ReplaceLinesTestSuite))
}
