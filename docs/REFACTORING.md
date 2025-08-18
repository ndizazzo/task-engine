# Refactoring Guide

This document explains how to migrate existing actions to the new DRY patterns introduced in the task-engine refactoring.

## Overview

The refactoring introduces two main patterns to reduce code duplication:

1. **New Action Style**: Generic constructor pattern using `common.BaseConstructor`
2. **New Parameter Style**: Centralized parameter resolution using `common.ParameterResolver` and `common.OutputBuilder`

## 1. Migrating to the New Action Style

### Before: Manual Constructor Pattern

```go
// Old style - manual constructor
func NewMyAction(logger *slog.Logger) *task_engine.Action[*MyAction] {
    action := &MyAction{
        BaseAction: task_engine.NewBaseAction(logger),
        // ... other fields
    }

    return &task_engine.Action[*MyAction]{
        Wrapped: action,
        Name:    "My Action",
        ID:      "my-action",
    }, nil
}
```

### After: Generic Constructor Pattern

```go
// New style - generic constructor
type MyActionConstructor struct {
    common.BaseConstructor[*MyAction]
}

func NewMyAction(logger *slog.Logger) *MyActionConstructor {
    return &MyActionConstructor{
        BaseConstructor: *common.NewBaseConstructor[*MyAction](logger),
    }
}

func (c *MyActionConstructor) WithParameters(
    param1 task_engine.ActionParameter,
    param2 task_engine.ActionParameter,
) (*task_engine.Action[*MyAction], error) {
    action := &MyAction{
        BaseAction: task_engine.NewBaseAction(c.GetLogger()),
        // ... other fields
        Param1: param1,
        Param2: param2,
    }

    return c.WrapAction(action, "My Action", "my-action"), nil
}
```

### Migration Steps

1. **Replace the constructor function** with a constructor struct
2. **Embed `common.BaseConstructor[T]`** where `T` is your action type
3. **Use `common.NewBaseConstructor[T]`** in the constructor
4. **Replace manual action wrapping** with `c.WrapAction()`
5. **Use `c.GetLogger()`** instead of passing logger directly

### Benefits

- **Consistent naming**: All constructors follow the same pattern
- **Reduced boilerplate**: No more manual action wrapping
- **Type safety**: Generic constraints ensure proper action types
- **Centralized logging**: Logger management is handled automatically

## 2. Migrating to the New Parameter Style

### Before: Manual Parameter Resolution

```go
// Old style - manual parameter resolution
func (a *MyAction) Execute(execCtx context.Context) error {
    // Extract GlobalContext
    globalCtx, ok := execCtx.Value("global_context").(task_engine.GlobalContext)
    if !ok {
        return fmt.Errorf("global context not found in execution context")
    }

    // Resolve parameters manually
    sourcePath, err := a.resolveStringParameter(globalCtx, a.SourcePathParam, "source path")
    if err != nil {
        return err
    }

    // ... more parameter resolution

    // Manual output building
    a.Output = map[string]interface{}{
        "success": true,
        "sourcePath": sourcePath,
        "result": result,
    }

    return nil
}

func (a *MyAction) GetOutput() interface{} {
    return map[string]interface{}{
        "success": true,
        "sourcePath": a.SourcePath,
        "result": a.Result,
    }
}
```

### After: Centralized Parameter Resolution

```go
// New style - embedded parameter resolution
type MyAction struct {
    task_engine.BaseAction
    common.ParameterResolver // Embed ParameterResolver
    common.OutputBuilder     // Embed OutputBuilder

    SourcePath string
    Result     string

    // Parameter-aware fields
    SourcePathParam task_engine.ActionParameter
    ResultParam     task_engine.ActionParameter
}

func NewMyAction(logger *slog.Logger) *MyActionConstructor {
    return &MyActionConstructor{
        BaseConstructor: *common.NewBaseConstructor[*MyAction](logger),
    }
}

func (c *MyActionConstructor) WithParameters(
    sourcePathParam task_engine.ActionParameter,
    resultParam task_engine.ActionParameter,
) (*task_engine.Action[*MyAction], error) {
    action := &MyAction{
        BaseAction:        task_engine.NewBaseAction(c.GetLogger()),
        ParameterResolver: *common.NewParameterResolver(c.GetLogger()), // Initialize resolver
        OutputBuilder:     *common.NewOutputBuilder(c.GetLogger()),     // Initialize builder
        SourcePathParam:   sourcePathParam,
        ResultParam:       resultParam,
    }

    return c.WrapAction(action, "My Action", "my-action"), nil
}

func (a *MyAction) Execute(execCtx context.Context) error {
    // Use embedded ParameterResolver methods
    sourcePath, err := a.ResolveStringParameter(execCtx, a.SourcePathParam, "source path")
    if err != nil {
        return err
    }

    result, err := a.ResolveStringParameter(execCtx, a.ResultParam, "result")
    if err != nil {
        return err
    }

    // ... action logic ...

    // Use embedded OutputBuilder
    a.Output = a.BuildStandardOutput(nil, true, map[string]interface{}{
        "sourcePath": sourcePath,
        "result":     result,
    })

    return nil
}

func (a *MyAction) GetOutput() interface{} {
    return a.BuildStandardOutput(nil, true, map[string]interface{}{
        "sourcePath": a.SourcePath,
        "result":     a.Result,
    })
}
```

### Migration Steps

1. **Embed the helper structs** in your action:

   ```go
   type MyAction struct {
       task_engine.BaseAction
       common.ParameterResolver
       common.OutputBuilder
       // ... your fields
   }
   ```

2. **Initialize them in the constructor**:

   ```go
   action := &MyAction{
       BaseAction:        task_engine.NewBaseAction(c.GetLogger()),
       ParameterResolver: *common.NewParameterResolver(c.GetLogger()),
       OutputBuilder:     *common.NewOutputBuilder(c.GetLogger()),
       // ... other fields
   }
   ```

3. **Replace manual parameter resolution** with embedded methods:

   - `a.ResolveStringParameter()` for strings
   - `a.ResolveBoolParameter()` for booleans
   - `a.ResolveIntParameter()` for integers
   - `a.ResolveStringSliceParameter()` for string slices
   - `a.ResolveMapParameter()` for maps
   - `a.ResolveParameter()` for generic types

4. **Replace manual output building** with embedded methods:
   - `a.BuildStandardOutput()` for standard outputs
   - `a.BuildOutputWithCount()` for slice results
   - `a.BuildSimpleOutput()` for simple outputs
   - `a.BuildErrorOutput()` for error outputs

### Available Parameter Resolution Methods

```go
// String parameters
a.ResolveStringParameter(execCtx, param, "parameter name")

// Boolean parameters
a.ResolveBoolParameter(execCtx, param, "parameter name")

// Integer parameters
a.ResolveIntParameter(execCtx, param, "parameter name")

// String slice parameters
a.ResolveStringSliceParameter(execCtx, param, "parameter name")

// Duration parameters
a.ResolveDurationParameter(execCtx, param, "parameter name")

// Map parameters
a.ResolveMapParameter(execCtx, param, "parameter name")

// Generic parameters (returns interface{})
a.ResolveParameter(execCtx, param, "parameter name")
```

### Available Output Building Methods

```go
// Standard output with success flag and custom data
a.BuildStandardOutput(err, success, data)

// Output with count for slice results
a.BuildOutputWithCount(items, err, success, data)

// Simple output (success + data only)
a.BuildSimpleOutput(success, data)

// Error output
a.BuildErrorOutput(err, data)

// Automatic output from struct fields using reflection
a.BuildStructOutput(data, err, success)
```

## 3. Complete Migration Example

Here's a complete example showing the before and after:

### Before Migration

```go
package example

import (
    "context"
    "fmt"
    "log/slog"
    "task_engine"
)

func NewExampleAction(logger *slog.Logger) *task_engine.Action[*ExampleAction] {
    action := &ExampleAction{
        BaseAction: task_engine.NewBaseAction(logger),
    }

    return &task_engine.Action[*ExampleAction]{
        Wrapped: action,
        Name:    "Example Action",
        ID:      "example-action",
    }, nil
}

type ExampleAction struct {
    task_engine.BaseAction
    SourcePath string
    TargetPath string
    SourcePathParam task_engine.ActionParameter
    TargetPathParam task_engine.ActionParameter
}

func (a *ExampleAction) Execute(execCtx context.Context) error {
    globalCtx, ok := execCtx.Value("global_context").(task_engine.GlobalContext)
    if !ok {
        return fmt.Errorf("global context not found in execution context")
    }

    sourcePath, err := a.resolveStringParameter(globalCtx, a.SourcePathParam, "source path")
    if err != nil {
        return err
    }

    targetPath, err := a.resolveStringParameter(globalCtx, a.TargetPathParam, "target path")
    if err != nil {
        return err
    }

    a.SourcePath = sourcePath
    a.TargetPath = targetPath

    // ... action logic ...

    a.Output = map[string]interface{}{
        "success": true,
        "sourcePath": sourcePath,
        "targetPath": targetPath,
    }

    return nil
}

func (a *ExampleAction) GetOutput() interface{} {
    return map[string]interface{}{
        "success": true,
        "sourcePath": a.SourcePath,
        "targetPath": a.TargetPath,
    }
}

func (a *ExampleAction) resolveStringParameter(globalCtx task_engine.GlobalContext, param task_engine.ActionParameter, name string) (string, error) {
    if param == nil {
        return "", fmt.Errorf("%s parameter is required", name)
    }

    value, err := param.Resolve(globalCtx)
    if err != nil {
        return "", fmt.Errorf("failed to resolve %s parameter: %w", name, err)
    }

    strValue, ok := value.(string)
    if !ok {
        return "", fmt.Errorf("%s parameter is not a string, got %T", name, value)
    }

    return strValue, nil
}
```

### After Migration

```go
package example

import (
    "context"
    "log/slog"
    "task_engine"
    "github.com/ndizazzo/task-engine/actions/common"
)

type ExampleActionConstructor struct {
    common.BaseConstructor[*ExampleAction]
}

func NewExampleAction(logger *slog.Logger) *ExampleActionConstructor {
    return &ExampleActionConstructor{
        BaseConstructor: *common.NewBaseConstructor[*ExampleAction](logger),
    }
}

func (c *ExampleActionConstructor) WithParameters(
    sourcePathParam task_engine.ActionParameter,
    targetPathParam task_engine.ActionParameter,
) (*task_engine.Action[*ExampleAction], error) {
    action := &ExampleAction{
        BaseAction:        task_engine.NewBaseAction(c.GetLogger()),
        ParameterResolver: *common.NewParameterResolver(c.GetLogger()),
        OutputBuilder:     *common.NewOutputBuilder(c.GetLogger()),
        SourcePathParam:   sourcePathParam,
        TargetPathParam:   targetPathParam,
    }

    return c.WrapAction(action, "Example Action", "example-action"), nil
}

type ExampleAction struct {
    task_engine.BaseAction
    common.ParameterResolver
    common.OutputBuilder

    SourcePath string
    TargetPath string

    SourcePathParam task_engine.ActionParameter
    TargetPathParam task_engine.ActionParameter
}

func (a *ExampleAction) Execute(execCtx context.Context) error {
    sourcePath, err := a.ResolveStringParameter(execCtx, a.SourcePathParam, "source path")
    if err != nil {
        return err
    }

    targetPath, err := a.ResolveStringParameter(execCtx, a.TargetPathParam, "target path")
    if err != nil {
        return err
    }

    a.SourcePath = sourcePath
    a.TargetPath = targetPath

    // ... action logic ...

    a.Output = a.BuildStandardOutput(nil, true, map[string]interface{}{
        "sourcePath": sourcePath,
        "targetPath": targetPath,
    })

    return nil
}

func (a *ExampleAction) GetOutput() interface{} {
    return a.BuildStandardOutput(nil, true, map[string]interface{}{
        "sourcePath": a.SourcePath,
        "targetPath": a.TargetPath,
    })
}
```

## 4. Testing Considerations

When migrating, you may need to update test expectations:

### Error Message Changes

The new `ParameterResolver` provides more consistent error messages:

```go
// Old error messages
"parameter is not a string, got int"
"parameter is not a boolean, got string"

// New error messages
"parameter resolved to non-string value"
"parameter resolved to non-boolean value"
```

### Test Updates

```go
// Update test assertions to match new error messages
suite.Contains(err.Error(), "parameter resolved to non-string value")
suite.Contains(err.Error(), "parameter resolved to non-boolean value")
```

## 5. Benefits of Migration

- **Reduced code duplication**: Common patterns are centralized
- **Consistent error handling**: Standardized parameter validation
- **Easier maintenance**: Changes to common logic affect all actions
- **Better testing**: Consistent behavior across all actions
- **Type safety**: Generic constraints prevent type errors
- **Standardized outputs**: Consistent output format across actions

## 6. Migration Checklist

- [ ] Update constructor to use `common.BaseConstructor`
- [ ] Embed `common.ParameterResolver` in action struct
- [ ] Embed `common.OutputBuilder` in action struct
- [ ] Initialize embedded structs in constructor
- [ ] Replace manual parameter resolution with embedded methods
- [ ] Replace manual output building with embedded methods
- [ ] Update tests to match new error messages
- [ ] Verify all tests pass
- [ ] Test parameter resolution with various types
- [ ] Test output generation

## 7. Common Pitfalls

1. **Forgetting to initialize embedded structs** in the constructor
2. **Not updating test error message expectations**
3. **Missing parameter type validation** in the new pattern
4. **Inconsistent output format** after migration

## 8. Getting Help

If you encounter issues during migration:

1. Check the existing migrated actions for examples
2. Review the `actions/common/` package for available methods
3. Run tests to identify specific issues
4. Ensure all embedded structs are properly initialized

The migration process is designed to be straightforward and maintain backward compatibility while providing significant improvements in code maintainability and consistency.
