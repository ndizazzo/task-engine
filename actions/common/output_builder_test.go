package common_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/stretchr/testify/suite"
)

type OutputBuilderTestSuite struct {
	suite.Suite
	builder *common.OutputBuilder
	logger  *slog.Logger
}

func (suite *OutputBuilderTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	suite.builder = common.NewOutputBuilder(suite.logger)
}

func (suite *OutputBuilderTestSuite) TestNewOutputBuilder() {
	ob := common.NewOutputBuilder(suite.logger)
	suite.NotNil(ob)
}

func (suite *OutputBuilderTestSuite) TestGetLogger() {
	logger := suite.builder.GetLogger()
	suite.NotNil(logger)
	suite.Equal(suite.logger, logger)
}

func (suite *OutputBuilderTestSuite) TestBuildStandardOutput_WithSuccess() {
	output := map[string]interface{}{"result": "value"}
	additionalFields := map[string]interface{}{"extra": "data"}

	result := suite.builder.BuildStandardOutput(output, true, additionalFields)

	suite.NotNil(result)
	suite.Equal(true, result["success"])
	suite.Equal(output, result["output"])
	suite.Equal("data", result["extra"])
}

func (suite *OutputBuilderTestSuite) TestBuildStandardOutput_WithFailure() {
	output := "error message"
	additionalFields := map[string]interface{}{}

	result := suite.builder.BuildStandardOutput(output, false, additionalFields)

	suite.Equal(false, result["success"])
	suite.Equal("error message", result["output"])
}

func (suite *OutputBuilderTestSuite) TestBuildStandardOutput_NoAdditionalFields() {
	result := suite.builder.BuildStandardOutput("test", true, nil)

	suite.Equal(true, result["success"])
	suite.Equal("test", result["output"])
	suite.Len(result, 2)
}

func (suite *OutputBuilderTestSuite) TestBuildStandardOutput_MultipleAdditionalFields() {
	additionalFields := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
		"field3": true,
		"field4": []string{"a", "b"},
	}

	result := suite.builder.BuildStandardOutput("output", true, additionalFields)

	suite.Equal("value1", result["field1"])
	suite.Equal(42, result["field2"])
	suite.Equal(true, result["field3"])
	suite.Equal([]string{"a", "b"}, result["field4"])
}

func (suite *OutputBuilderTestSuite) TestBuildOutputFromStruct_WithBasicStruct() {
	type TestStruct struct {
		Name  string
		Age   int
		Email string
	}

	testObj := TestStruct{
		Name:  "John",
		Age:   30,
		Email: "john@example.com",
	}

	result := suite.builder.BuildOutputFromStruct(testObj, true, nil)

	suite.Equal(true, result["success"])
	suite.Equal("John", result["name"])
	suite.Equal(30, result["age"])
	suite.Equal("john@example.com", result["email"])
}

func (suite *OutputBuilderTestSuite) TestBuildOutputFromStruct_WithPointerStruct() {
	type TestStruct struct {
		Name string
		Age  int
	}

	testObj := &TestStruct{
		Name: "Jane",
		Age:  25,
	}

	result := suite.builder.BuildOutputFromStruct(testObj, true, nil)

	suite.Equal(true, result["success"])
	suite.Equal("Jane", result["name"])
	suite.Equal(25, result["age"])
}

func (suite *OutputBuilderTestSuite) TestBuildOutputFromStruct_ExcludeFields() {
	type TestStruct struct {
		Name     string
		Password string
		Email    string
	}

	testObj := TestStruct{
		Name:     "John",
		Password: "secret123",
		Email:    "john@example.com",
	}

	result := suite.builder.BuildOutputFromStruct(testObj, true, []string{"Password"})

	suite.Equal("John", result["name"])
	suite.Equal("john@example.com", result["email"])
	suite.NotContains(result, "Password")
	suite.NotContains(result, "password")
}

func (suite *OutputBuilderTestSuite) TestBuildOutputFromStruct_WithZeroValues() {
	type TestStruct struct {
		Name   string
		Age    int
		Email  string
		Active bool
	}

	testObj := TestStruct{
		Name:   "John",
		Age:    0,     // zero value - should be skipped
		Email:  "",    // zero value - should be skipped
		Active: false, // zero value - should be skipped
	}

	result := suite.builder.BuildOutputFromStruct(testObj, true, nil)

	suite.Equal("John", result["name"])
	suite.NotContains(result, "age")
	suite.NotContains(result, "email")
	suite.NotContains(result, "active")
}

func (suite *OutputBuilderTestSuite) TestBuildOutputFromStruct_NonStruct() {
	result := suite.builder.BuildOutputFromStruct("string value", true, nil)

	suite.Equal(true, result["success"])
	suite.Len(result, 1)
}

func (suite *OutputBuilderTestSuite) TestBuildOutputFromStruct_SkipsUnexportedFields() {
	type TestStruct struct {
		Public  string
		private string // unexported
	}

	testObj := TestStruct{
		Public:  "visible",
		private: "hidden",
	}

	result := suite.builder.BuildOutputFromStruct(testObj, true, nil)

	suite.Equal("visible", result["public"])
	suite.NotContains(result, "private")
}

func (suite *OutputBuilderTestSuite) TestBuildSimpleOutput_WithMessage() {
	result := suite.builder.BuildSimpleOutput(true, "Operation completed successfully")

	suite.Equal(true, result["success"])
	suite.Equal("Operation completed successfully", result["message"])
	suite.Len(result, 2)
}

func (suite *OutputBuilderTestSuite) TestBuildSimpleOutput_WithoutMessage() {
	result := suite.builder.BuildSimpleOutput(true, "")

	suite.Equal(true, result["success"])
	suite.NotContains(result, "message")
	suite.Len(result, 1)
}

func (suite *OutputBuilderTestSuite) TestBuildSimpleOutput_WithFailure() {
	result := suite.builder.BuildSimpleOutput(false, "Operation failed")

	suite.Equal(false, result["success"])
	suite.Equal("Operation failed", result["message"])
}

func (suite *OutputBuilderTestSuite) TestBuildErrorOutput_WithError() {
	result := suite.builder.BuildErrorOutput("Custom error message", nil)

	suite.Equal(false, result["success"])
	suite.Equal("Custom error message", result["error"])
	suite.Len(result, 2)
}

func (suite *OutputBuilderTestSuite) TestBuildErrorOutput_WithAdditionalFields() {
	additionalFields := map[string]interface{}{
		"errorCode": "ERR_001",
		"timestamp": "2024-01-01T00:00:00Z",
	}

	result := suite.builder.BuildErrorOutput("Something went wrong", additionalFields)

	suite.Equal(false, result["success"])
	suite.Equal("Something went wrong", result["error"])
	suite.Equal("ERR_001", result["errorCode"])
	suite.Equal("2024-01-01T00:00:00Z", result["timestamp"])
}

func (suite *OutputBuilderTestSuite) TestBuildErrorOutput_WithNilError() {
	result := suite.builder.BuildErrorOutput(nil, nil)

	suite.Equal(false, result["success"])
	suite.Nil(result["error"])
}

func (suite *OutputBuilderTestSuite) TestBuildOutputWithCount_WithSliceItems() {
	items := []string{"item1", "item2", "item3"}

	result := suite.builder.BuildOutputWithCount(items, true, nil)

	suite.Equal(true, result["success"])
	suite.Equal(items, result["output"])
	suite.Equal(3, result["count"])
}

func (suite *OutputBuilderTestSuite) TestBuildOutputWithCount_WithEmptySlice() {
	items := []string{}

	result := suite.builder.BuildOutputWithCount(items, true, nil)

	suite.Equal(true, result["success"])
	suite.Equal(0, result["count"])
}

func (suite *OutputBuilderTestSuite) TestBuildOutputWithCount_WithNonSliceItems() {
	result := suite.builder.BuildOutputWithCount("not a slice", true, nil)

	suite.Equal(true, result["success"])
	suite.Equal("not a slice", result["output"])
	suite.NotContains(result, "count")
}

func (suite *OutputBuilderTestSuite) TestBuildOutputWithCount_WithNilItems() {
	result := suite.builder.BuildOutputWithCount(nil, true, nil)

	suite.Equal(true, result["success"])
	suite.Nil(result["output"])
	suite.NotContains(result, "count")
}

func (suite *OutputBuilderTestSuite) TestBuildOutputWithCount_WithAdditionalFields() {
	items := []int{1, 2, 3, 4, 5}
	additionalFields := map[string]interface{}{
		"page":     1,
		"pageSize": 10,
	}

	result := suite.builder.BuildOutputWithCount(items, true, additionalFields)

	suite.Equal(5, result["count"])
	suite.Equal(1, result["page"])
	suite.Equal(10, result["pageSize"])
}

func (suite *OutputBuilderTestSuite) TestBuildOutputWithCount_WithComplexSlice() {
	type Item struct {
		ID   int
		Name string
	}

	items := []Item{
		{ID: 1, Name: "Item1"},
		{ID: 2, Name: "Item2"},
	}

	result := suite.builder.BuildOutputWithCount(items, true, nil)

	suite.Equal(2, result["count"])
	suite.Equal(items, result["output"])
}

func (suite *OutputBuilderTestSuite) TestBuildOutputFromStruct_WithSliceField() {
	type TestStruct struct {
		Name  string
		Items []string
	}

	testObj := TestStruct{
		Name:  "John",
		Items: []string{"a", "b", "c"},
	}

	result := suite.builder.BuildOutputFromStruct(testObj, true, nil)

	suite.Equal("John", result["name"])
	suite.Equal([]string{"a", "b", "c"}, result["items"])
}

func (suite *OutputBuilderTestSuite) TestBuildOutputFromStruct_CamelCase() {
	type TestStruct struct {
		FirstName    string
		LastName     string
		EmailAddress string
	}

	testObj := TestStruct{
		FirstName:    "John",
		LastName:     "Doe",
		EmailAddress: "john@example.com",
	}

	result := suite.builder.BuildOutputFromStruct(testObj, true, nil)

	suite.Equal("John", result["firstName"])
	suite.Equal("Doe", result["lastName"])
	suite.Equal("john@example.com", result["emailAddress"])
}

func TestOutputBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(OutputBuilderTestSuite))
}
