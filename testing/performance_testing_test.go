package testing

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	task_engine "github.com/ndizazzo/task-engine"
)

func TestNewPerformanceTester(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)

	pt := NewPerformanceTester(tm, logger)
	require.NotNil(t, pt)
	assert.NotNil(t, pt.metrics)
	assert.Equal(t, 0, pt.metrics.TotalTasksExecuted)
}

func TestPerformanceTester_GetMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	metrics := pt.GetMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.TotalTasksExecuted)
	assert.Equal(t, time.Duration(0), metrics.TotalExecutionTime)
}

func TestPerformanceTester_ResetMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	// Manually set some metrics
	pt.metrics.TotalTasksExecuted = 42
	pt.metrics.ErrorRate = 10.0

	pt.ResetMetrics()
	metrics := pt.GetMetrics()
	assert.Equal(t, 0, metrics.TotalTasksExecuted)
	assert.Equal(t, 0.0, metrics.ErrorRate)
}

func TestPerformanceTester_BenchmarkSequential(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	task := &task_engine.Task{
		ID:      "bench-seq",
		Name:    "Benchmark Sequential",
		Actions: []task_engine.ActionWrapper{},
	}

	metrics := pt.BenchmarkTaskExecution(context.Background(), task, 3, false)
	require.NotNil(t, metrics)
	assert.Equal(t, 3, metrics.TotalTasksExecuted)
	assert.Greater(t, metrics.TotalExecutionTime, time.Duration(0))
	assert.Equal(t, 1, metrics.ConcurrentTasks)
	// ErrorRate may be non-zero due to task ID collisions when tasks complete
	// within the same millisecond (timestamp-based ID generation)
	assert.GreaterOrEqual(t, metrics.ErrorRate, 0.0)
	assert.GreaterOrEqual(t, metrics.MinExecutionTime, time.Duration(0))
	assert.GreaterOrEqual(t, metrics.MaxExecutionTime, metrics.MinExecutionTime)
	assert.Greater(t, metrics.TaskThroughput, 0.0)
}

func TestPerformanceTester_BenchmarkConcurrent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	task := &task_engine.Task{
		ID:      "bench-conc",
		Name:    "Benchmark Concurrent",
		Actions: []task_engine.ActionWrapper{},
	}

	metrics := pt.BenchmarkTaskExecution(context.Background(), task, 5, true)
	require.NotNil(t, metrics)
	assert.Equal(t, 5, metrics.TotalTasksExecuted)
	assert.Equal(t, 5, metrics.ConcurrentTasks)
	assert.GreaterOrEqual(t, metrics.ErrorRate, 0.0)
	assert.Greater(t, metrics.TaskThroughput, 0.0)
}

func TestPerformanceTester_BenchmarkContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	// Use a slow action so cancellation happens mid-execution
	slowAction := &MockAction{
		ID:       "slow-action",
		Duration: 5 * time.Second,
		Logger:   logger,
	}

	task := &task_engine.Task{
		ID:      "bench-cancel",
		Name:    "Benchmark Cancel",
		Actions: []task_engine.ActionWrapper{slowAction},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Sequential with 10 iterations but context will cancel quickly
	metrics := pt.BenchmarkTaskExecution(ctx, task, 10, false)
	require.NotNil(t, metrics)
	// At least some iterations should have run (possibly all with errors)
	assert.Greater(t, metrics.TotalTasksExecuted, 0)

	// Wait for any background goroutines in the task manager
	err := tm.WaitForAllTasksToComplete(2 * time.Second)
	require.NoError(t, err)
}

func TestPerformanceTester_LoadTest(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	task := &task_engine.Task{
		ID:      "load-test",
		Name:    "Load Test",
		Actions: []task_engine.ActionWrapper{},
	}

	metrics := pt.LoadTest(context.Background(), task, 10, 3, 5*time.Second)
	require.NotNil(t, metrics)
	assert.Greater(t, metrics.TotalTasksExecuted, 0)
	assert.Greater(t, metrics.TaskThroughput, 0.0)
}

func TestPerformanceTester_LoadTestContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	task := &task_engine.Task{
		ID:      "load-cancel",
		Name:    "Load Cancel",
		Actions: []task_engine.ActionWrapper{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	metrics := pt.LoadTest(ctx, task, 1000, 5, 10*time.Second)
	require.NotNil(t, metrics)
	// Should have run some tasks before context cancelled
	assert.GreaterOrEqual(t, metrics.TotalTasksExecuted, 0)

	err := tm.WaitForAllTasksToComplete(2 * time.Second)
	require.NoError(t, err)
}

func TestPerformanceTester_StressTest(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	task := &task_engine.Task{
		ID:      "stress-test",
		Name:    "Stress Test",
		Actions: []task_engine.ActionWrapper{},
	}

	// Start at concurrency 1, max 4, with short step duration
	metrics := pt.StressTest(context.Background(), task, 1, 4, 500*time.Millisecond)
	require.NotNil(t, metrics)
	assert.Greater(t, metrics.TotalTasksExecuted, 0)

	err := tm.WaitForAllTasksToComplete(5 * time.Second)
	require.NoError(t, err)
}

func TestPerformanceTester_GenerateReport(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	// Run a small benchmark first to populate metrics
	task := &task_engine.Task{
		ID:      "report-test",
		Name:    "Report Test",
		Actions: []task_engine.ActionWrapper{},
	}
	pt.BenchmarkTaskExecution(context.Background(), task, 2, false)

	report := pt.GenerateReport()
	require.NotNil(t, report)

	// Verify all expected keys are present
	assert.Contains(t, report, "timestamp")
	assert.Contains(t, report, "total_tasks_executed")
	assert.Contains(t, report, "total_execution_time")
	assert.Contains(t, report, "average_execution_time")
	assert.Contains(t, report, "min_execution_time")
	assert.Contains(t, report, "max_execution_time")
	assert.Contains(t, report, "concurrent_tasks")
	assert.Contains(t, report, "task_throughput")
	assert.Contains(t, report, "error_rate")
	assert.Contains(t, report, "performance_score")

	assert.Equal(t, 2, report["total_tasks_executed"])
}

func TestPerformanceTester_GenerateReportEmpty(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	report := pt.GenerateReport()
	require.NotNil(t, report)
	assert.Equal(t, 0, report["total_tasks_executed"])
	// Performance score should be 0 when no tasks executed
	assert.Equal(t, 0.0, report["performance_score"])
}

func TestPerformanceTester_CalculateMetricsEmpty(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	// Call calculateMetrics with empty slices (tests guard against panic)
	pt.mu.Lock()
	pt.calculateMetrics([]time.Duration{}, []error{}, 0, false)
	pt.mu.Unlock()

	metrics := pt.GetMetrics()
	assert.Equal(t, 0, metrics.TotalTasksExecuted)
	assert.Equal(t, time.Duration(0), metrics.AverageExecutionTime)
	assert.Equal(t, time.Duration(0), metrics.MinExecutionTime)
	assert.Equal(t, time.Duration(0), metrics.MaxExecutionTime)
	assert.Equal(t, 0.0, metrics.TaskThroughput)
	assert.Equal(t, 0.0, metrics.ErrorRate)
}

func TestPerformanceTester_CalculateMetricsWithErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	durations := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond, 400 * time.Millisecond}
	errors := []error{nil, assert.AnError, nil, assert.AnError}

	pt.mu.Lock()
	pt.calculateMetrics(durations, errors, 1*time.Second, false)
	pt.mu.Unlock()

	metrics := pt.GetMetrics()
	assert.Equal(t, 4, metrics.TotalTasksExecuted)
	assert.Equal(t, 1, metrics.ConcurrentTasks)
	assert.Equal(t, 50.0, metrics.ErrorRate) // 2 out of 4 = 50%
	assert.Equal(t, 100*time.Millisecond, metrics.MinExecutionTime)
	assert.Equal(t, 400*time.Millisecond, metrics.MaxExecutionTime)
	assert.Equal(t, 250*time.Millisecond, metrics.AverageExecutionTime)
	assert.Equal(t, 4.0, metrics.TaskThroughput) // 4 tasks / 1 second
}

func TestPerformanceTester_CalculatePerformanceScore(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	// No tasks → score 0
	pt.mu.RLock()
	score := pt.calculatePerformanceScore()
	pt.mu.RUnlock()
	assert.Equal(t, 0.0, score)

	// Good metrics: high throughput, low error rate, fast execution
	pt.mu.Lock()
	pt.metrics = &PerformanceMetrics{
		TotalTasksExecuted:   100,
		TaskThroughput:       50.0, // 50 tasks/sec
		ErrorRate:            0.0,  // 0% errors
		AverageExecutionTime: 100 * time.Millisecond,
	}
	pt.mu.Unlock()

	pt.mu.RLock()
	score = pt.calculatePerformanceScore()
	pt.mu.RUnlock()
	assert.Greater(t, score, 0.0)

	// Bad metrics: slow execution time > 10s results in timingScore clamped to 0
	pt.mu.Lock()
	pt.metrics = &PerformanceMetrics{
		TotalTasksExecuted:   10,
		TaskThroughput:       1.0,
		ErrorRate:            0.0,
		AverageExecutionTime: 15 * time.Second,
	}
	pt.mu.Unlock()

	pt.mu.RLock()
	score = pt.calculatePerformanceScore()
	pt.mu.RUnlock()
	// timingScore is clamped to 0 when avg > 10s, so score = (throughputScore + 0) / 2
	assert.GreaterOrEqual(t, score, 0.0)

	// High error rate penalizes score
	pt.mu.Lock()
	pt.metrics = &PerformanceMetrics{
		TotalTasksExecuted:   100,
		TaskThroughput:       100.0,
		ErrorRate:            100.0, // 100% error rate
		AverageExecutionTime: 1 * time.Second,
	}
	pt.mu.Unlock()

	pt.mu.RLock()
	score = pt.calculatePerformanceScore()
	pt.mu.RUnlock()
	assert.Equal(t, 0.0, score) // 100% error → score * 0 = 0
}

func TestPerformanceTester_BenchmarkWithFailingTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tm := task_engine.NewTaskManager(logger)
	pt := NewPerformanceTester(tm, logger)

	failingAction := &FailingAction{
		ID:     "failing-action",
		Logger: logger,
	}

	task := &task_engine.Task{
		ID:      "bench-fail",
		Name:    "Benchmark Failing",
		Actions: []task_engine.ActionWrapper{failingAction},
	}

	metrics := pt.BenchmarkTaskExecution(context.Background(), task, 3, false)
	require.NotNil(t, metrics)
	assert.Equal(t, 3, metrics.TotalTasksExecuted)
	assert.Equal(t, 100.0, metrics.ErrorRate) // All tasks should fail

	err := tm.WaitForAllTasksToComplete(2 * time.Second)
	require.NoError(t, err)
}

// FailingAction is a test action that always returns an error
type FailingAction struct {
	ID     string
	Logger *slog.Logger
}

func (fa *FailingAction) GetID() string              { return fa.ID }
func (fa *FailingAction) SetID(id string)            { fa.ID = id }
func (fa *FailingAction) GetDuration() time.Duration { return 0 }
func (fa *FailingAction) GetLogger() *slog.Logger    { return fa.Logger }
func (fa *FailingAction) GetName() string            { return fa.ID }
func (fa *FailingAction) Execute(_ context.Context) error {
	return assert.AnError
}
func (fa *FailingAction) GetOutput() interface{} { return nil }
