package file_test

import (
	"testing"

	"github.com/ndizazzo/task-engine/actions/file"
	"github.com/stretchr/testify/suite"
)

type PathValidationTestSuite struct {
	suite.Suite
}

func (suite *PathValidationTestSuite) TestValidatePath() {
	tests := []struct {
		name          string
		path          string
		pathType      string
		expectError   bool
		errorContains string
	}{
		{
			name:          "empty path",
			path:          "",
			pathType:      "test",
			expectError:   true,
			errorContains: "test path cannot be empty",
		},
		{
			name:        "absolute path unix",
			path:        "/usr/local/bin",
			pathType:    "source",
			expectError: false,
		},
		{
			name:        "absolute path windows",
			path:        "C:\\Users\\test",
			pathType:    "destination",
			expectError: false,
		},
		{
			name:        "relative path with dot",
			path:        "./config/settings.json",
			pathType:    "source",
			expectError: false,
		},
		{
			name:        "relative path with double dot",
			path:        "../config/settings.json",
			pathType:    "source",
			expectError: false,
		},
		{
			name:        "simple relative path",
			path:        "config/settings.json",
			pathType:    "source",
			expectError: false,
		},
		{
			name:          "path traversal attack",
			path:          "../../etc/passwd",
			pathType:      "source",
			expectError:   true,
			errorContains: "contains potentially dangerous path traversal",
		},
		{
			name:          "complex path traversal",
			path:          "config/../../../etc/passwd",
			pathType:      "destination",
			expectError:   true,
			errorContains: "contains potentially dangerous path traversal",
		},
		{
			name:        "single file name",
			path:        "config.json",
			pathType:    "source",
			expectError: false,
		},
		{
			name:        "path with spaces",
			path:        "./config/my settings.json",
			pathType:    "source",
			expectError: false,
		},
		{
			name:        "current directory",
			path:        ".",
			pathType:    "source",
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := file.ValidatePath(tt.path, tt.pathType)
			if tt.expectError {
				suite.Error(err, "Expected error for path: %s", tt.path)
				if tt.errorContains != "" {
					suite.Contains(err.Error(), tt.errorContains)
				}
			} else {
				suite.NoError(err, "Expected no error for path: %s", tt.path)
			}
		})
	}
}

func (suite *PathValidationTestSuite) TestValidateSourcePath() {
	tests := []struct {
		name          string
		path          string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid source path",
			path:        "./source/data.txt",
			expectError: false,
		},
		{
			name:          "empty source path",
			path:          "",
			expectError:   true,
			errorContains: "source path cannot be empty",
		},
		{
			name:          "dangerous source path",
			path:          "../../../etc/passwd",
			expectError:   true,
			errorContains: "potentially dangerous path traversal",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := file.ValidateSourcePath(tt.path)
			if tt.expectError {
				suite.Error(err)
				if tt.errorContains != "" {
					suite.Contains(err.Error(), tt.errorContains)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *PathValidationTestSuite) TestValidateDestinationPath() {
	tests := []struct {
		name          string
		path          string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid destination path",
			path:        "./dest/output.txt",
			expectError: false,
		},
		{
			name:          "empty destination path",
			path:          "",
			expectError:   true,
			errorContains: "destination path cannot be empty",
		},
		{
			name:          "dangerous destination path",
			path:          "../../root/.ssh/authorized_keys",
			expectError:   true,
			errorContains: "potentially dangerous path traversal",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := file.ValidateDestinationPath(tt.path)
			if tt.expectError {
				suite.Error(err)
				if tt.errorContains != "" {
					suite.Contains(err.Error(), tt.errorContains)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *PathValidationTestSuite) TestSanitizePath() {
	tests := []struct {
		name          string
		path          string
		expectedPath  string
		expectError   bool
		errorContains string
	}{
		{
			name:          "empty path",
			path:          "",
			expectError:   true,
			errorContains: "path cannot be empty",
		},
		{
			name:         "absolute path",
			path:         "/usr/local/bin",
			expectedPath: "/usr/local/bin",
			expectError:  false,
		},
		{
			name:         "absolute path with redundant separators",
			path:         "/usr//local///bin",
			expectedPath: "/usr/local/bin",
			expectError:  false,
		},
		{
			name:         "relative path with dot",
			path:         "./config/settings.json",
			expectedPath: "config/settings.json",
			expectError:  false,
		},
		{
			name:         "relative path with current dir references",
			path:         "config/./settings.json",
			expectedPath: "config/settings.json",
			expectError:  false,
		},
		{
			name:         "simple relative path",
			path:         "config/settings.json",
			expectedPath: "config/settings.json",
			expectError:  false,
		},
		{
			name:          "path traversal attack",
			path:          "../../etc/passwd",
			expectError:   true,
			errorContains: "contains potentially dangerous path traversal",
		},
		{
			name:          "complex path traversal",
			path:          "config/../../../etc/passwd",
			expectError:   true,
			errorContains: "contains potentially dangerous path traversal",
		},
		{
			name:         "single file name",
			path:         "config.json",
			expectedPath: "config.json",
			expectError:  false,
		},
		{
			name:         "current directory",
			path:         ".",
			expectedPath: ".",
			expectError:  false,
		},
		{
			name:         "parent directory prefix allowed",
			path:         "../config/settings.json",
			expectedPath: "../config/settings.json",
			expectError:  false,
		},
		{
			name:         "multiple redundant slashes",
			path:         "config///test//file.txt",
			expectedPath: "config/test/file.txt",
			expectError:  false,
		},
		{
			name:         "trailing slash",
			path:         "config/directory/",
			expectedPath: "config/directory",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := file.SanitizePath(tt.path)
			if tt.expectError {
				suite.Error(err, "Expected error for path: %s", tt.path)
				if tt.errorContains != "" {
					suite.Contains(err.Error(), tt.errorContains)
				}
				suite.Empty(result, "Result should be empty on error")
			} else {
				suite.NoError(err, "Expected no error for path: %s", tt.path)
				suite.Equal(tt.expectedPath, result, "Sanitized path should match expected")
			}
		})
	}
}

func (suite *PathValidationTestSuite) TestSanitizePathEdgeCases() {
	tests := []struct {
		name         string
		path         string
		expectedPath string
		expectError  bool
	}{
		{
			name:         "path with spaces",
			path:         "./config/my file.txt",
			expectedPath: "config/my file.txt",
			expectError:  false,
		},
		{
			name:         "path with unicode characters",
			path:         "./config/файл.txt",
			expectedPath: "config/файл.txt",
			expectError:  false,
		},
		{
			name:         "Windows-style path",
			path:         ".\\config\\settings.json",
			expectedPath: "config/settings.json", // filepath.Clean normalizes separators
			expectError:  false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := file.SanitizePath(tt.path)
			if tt.expectError {
				suite.Error(err)
			} else {
				suite.NoError(err)
				suite.Equal(tt.expectedPath, result)
			}
		})
	}
}

func (suite *PathValidationTestSuite) TestPathValidationSecurity() {
	dangerousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"config/../../../etc/shadow",
		"./config/../../.ssh/id_rsa",
		"data/../../../../../../root/.bashrc",
		"uploads/../../../var/www/html/shell.php",
	}

	for _, dangerousPath := range dangerousPaths {
		suite.Run("dangerous_path_"+dangerousPath, func() {
			err := file.ValidatePath(dangerousPath, "test")
			suite.Error(err, "Should reject dangerous path: %s", dangerousPath)
			suite.Contains(err.Error(), "potentially dangerous path traversal")

			_, err = file.SanitizePath(dangerousPath)
			suite.Error(err, "SanitizePath should reject dangerous path: %s", dangerousPath)
		})
	}
}

func (suite *PathValidationTestSuite) TestPathValidationAllowedPaths() {
	allowedPaths := []string{
		"/absolute/path/to/file.txt",
		"./relative/path/file.txt",
		"../parent/directory/file.txt",
		"simple_file.txt",
		"config/settings.json",
		"data/input/large_file.dat",
		"./config",
		"../config",
		".",
	}

	for _, allowedPath := range allowedPaths {
		suite.Run("allowed_path_"+allowedPath, func() {
			err := file.ValidatePath(allowedPath, "test")
			suite.NoError(err, "Should allow safe path: %s", allowedPath)

			result, err := file.SanitizePath(allowedPath)
			suite.NoError(err, "SanitizePath should allow safe path: %s", allowedPath)
			suite.NotEmpty(result)
		})
	}
}

func TestPathValidationTestSuite(t *testing.T) {
	suite.Run(t, new(PathValidationTestSuite))
}
