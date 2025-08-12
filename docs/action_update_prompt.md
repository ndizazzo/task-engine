# Action Update Prompt

You must update all actions in the task-engine codebase to follow the single builder pattern documented in docs/passing.md. Apply these rules strictly:

## Required Implementation Pattern

1. **Single Constructor**: Only implement NewActionName(logger *slog.Logger) *ActionName - returns the action struct directly
2. **Single Action Struct**: ActionName struct contains BaseAction and parameter fields - no separate builder struct needed
3. **WithParameters Method**: Method on action struct that returns (*task_engine.Action[*ActionName], error) - always include error return
4. **Parameter Validation**: Check parameters are not nil in WithParameters, return descriptive errors
5. **Parameter-Only Action Struct**: Action struct contains ONLY task_engine.BaseAction, ActionParameter fields, and execution result fields (for GetOutput) - no static input fields
   - Input fields must be ActionParameter types (resolved at runtime)
   - Result fields (string, int, etc.) are allowed for storing execution results returned by GetOutput()
6. **Runtime Resolution**: All parameters resolved in Execute() method using globalContext
7. **Type Checking**: Validate resolved parameter types with meaningful error messages
8. **Error Wrapping**: Wrap parameter resolution errors with context using fmt.Errorf

## Prohibited Patterns

1. **No Legacy Constructors**: Remove all constructors that take static parameters directly
2. **No Static Input Fields**: Action structs cannot have string, int, []string or other static input value fields - use ActionParameter instead. Result fields for GetOutput() are allowed.
3. **No Builder Suffix**: Constructor names must be NewActionName, not NewActionNameBuilder
4. **No Separate Builder Structs**: Use single action struct with WithParameters method, no separate builder struct needed
5. **No WithParameters Without Error**: WithParameters must always return error as second value
6. **No Dual-Mode Structs**: No mixing static fields with parameter fields
7. **No Custom Builder Methods**: No WithDescription, WithTimeout, etc - only WithParameters
8. **No Helper Constructors**: Do not add NewActionNameWithStatic or similar convenience functions

## Test Update Requirements

1. **Use Builder Pattern**: Replace all direct constructor calls with NewActionName(logger).WithParameters(params)
2. **Handle Errors**: Always use suite.Require().NoError(err) after WithParameters calls
3. **Parameter Types**: Wrap all test values in task_engine.StaticParameter{Value: value}
4. **Error Variable Names**: Use execErr for Execute() errors to avoid shadowing
5. **GlobalContext Testing**: Include tests with ActionOutputParameter to verify parameter resolution

## Implementation Steps

1. Update action struct to remove all static fields, keep only ActionParameter fields
2. Update builder struct to contain only logger field
3. Remove Builder suffix from constructor function name
4. Add error return to WithParameters method with nil parameter validation
5. Update Execute method to resolve all parameters at runtime using local variables
6. Remove any legacy constructors or helper functions
7. Update all test files to use new pattern with proper error handling
8. Run tests after each action update to ensure correctness

Apply these changes systematically to each action, updating both the action file and its corresponding test file before moving to the next action. Always verify tests pass after each update.
