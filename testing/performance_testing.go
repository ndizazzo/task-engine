package testing

import (
	"context"
	"log/slog"
	"sync"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
)

// PerformanceMetrics holds performance-related data
type PerformanceMetrics struct {
	TotalTasksExecuted   int
	TotalExecutionTime   time.Duration
	AverageExecutionTime time.Duration
	MinExecutionTime     time.Duration
	MaxExecutionTime     time.Duration
	ConcurrentTasks      int
	MemoryUsage          uint64  // in bytes
	CPUUsage             float64 // percentage
	TaskThroughput       float64 // tasks per second
	ErrorRate            float64 // percentage of failed tasks
}

// PerformanceTester provides performance testing capabilities
type PerformanceTester struct {
	taskManager task_engine.TaskManagerInterface
	logger      *slog.Logger
	metrics     *PerformanceMetrics
	mu          sync.RWMutex
}

// NewPerformanceTester creates a new performance tester
func NewPerformanceTester(taskManager task_engine.TaskManagerInterface, logger *slog.Logger) *PerformanceTester {
	return &PerformanceTester{
		taskManager: taskManager,
		logger:      logger,
		metrics:     &PerformanceMetrics{},
	}
}

// BenchmarkTaskExecution runs a benchmark test for task execution
func (pt *PerformanceTester) BenchmarkTaskExecution(
	ctx context.Context,
	task *task_engine.Task,
	iterations int,
	concurrent bool,
) *PerformanceMetrics {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.logger.Info("Starting benchmark",
		"taskID", task.ID,
		"iterations", iterations,
		"concurrent", concurrent)

	startTime := time.Now()
	var wg sync.WaitGroup
	executionTimes := make([]time.Duration, iterations)
	errors := make([]error, iterations)

	if concurrent {
		// Run tasks concurrently
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				execTime, err := pt.executeSingleTask(ctx, task)
				executionTimes[index] = execTime
				errors[index] = err
			}(i)
		}
		wg.Wait()
	} else {
		// Run tasks sequentially
		for i := 0; i < iterations; i++ {
			execTime, err := pt.executeSingleTask(ctx, task)
			executionTimes[i] = execTime
			errors[i] = err
		}
	}

	totalTime := time.Since(startTime)
	pt.calculateMetrics(executionTimes, errors, totalTime, concurrent)

	pt.logger.Info("Benchmark completed",
		"totalTime", totalTime,
		"averageTime", pt.metrics.AverageExecutionTime)

	return pt.metrics
}

// executeSingleTask executes a single task and measures its execution time
func (pt *PerformanceTester) executeSingleTask(ctx context.Context, task *task_engine.Task) (time.Duration, error) {
	startTime := time.Now()

	// Create a copy of the task to avoid conflicts
	taskCopy := &task_engine.Task{
		ID:      task.ID + "_" + time.Now().Format("20060102150405"),
		Name:    task.Name,
		Actions: task.Actions,
		Logger:  pt.logger,
	}

	err := pt.taskManager.AddTask(taskCopy)
	if err != nil {
		return 0, err
	}

	err = pt.taskManager.RunTask(taskCopy.ID)
	if err != nil {
		return 0, err
	}

	// Wait for task completion or context cancellation
	select {
	case <-ctx.Done():
		// Stop the task when context is cancelled
		if stopErr := pt.taskManager.StopTask(taskCopy.ID); stopErr != nil {
			pt.logger.Warn("Failed to stop task during context cancellation",
				"taskID", taskCopy.ID,
				"error", stopErr)
		}
		return time.Since(startTime), ctx.Err()
	default:
		// Simple wait - in a real implementation, you might want to poll the task status
		time.Sleep(100 * time.Millisecond)
	}

	executionTime := time.Since(startTime)
	return executionTime, nil
}

// calculateMetrics calculates performance metrics from execution data
func (pt *PerformanceTester) calculateMetrics(
	executionTimes []time.Duration,
	errors []error,
	totalTime time.Duration,
	concurrent bool,
) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.metrics.TotalTasksExecuted = len(executionTimes)
	pt.metrics.TotalExecutionTime = totalTime
	pt.metrics.ConcurrentTasks = 1
	if concurrent {
		pt.metrics.ConcurrentTasks = len(executionTimes)
	}

	// Calculate timing metrics
	var totalExecTime time.Duration
	minTime := executionTimes[0]
	maxTime := executionTimes[0]

	for _, execTime := range executionTimes {
		totalExecTime += execTime
		if execTime < minTime {
			minTime = execTime
		}
		if execTime > maxTime {
			maxTime = execTime
		}
	}

	pt.metrics.AverageExecutionTime = totalExecTime / time.Duration(len(executionTimes))
	pt.metrics.MinExecutionTime = minTime
	pt.metrics.MaxExecutionTime = maxTime

	// Calculate throughput
	if totalTime > 0 {
		pt.metrics.TaskThroughput = float64(len(executionTimes)) / totalTime.Seconds()
	}

	// Calculate error rate
	errorCount := 0
	for _, err := range errors {
		if err != nil {
			errorCount++
		}
	}
	if len(errors) > 0 {
		pt.metrics.ErrorRate = float64(errorCount) / float64(len(errors)) * 100
	}
}

// LoadTest simulates high-load scenarios
func (pt *PerformanceTester) LoadTest(
	ctx context.Context,
	task *task_engine.Task,
	totalTasks int,
	concurrentLimit int,
	duration time.Duration,
) *PerformanceMetrics {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.logger.Info("Starting load test",
		"totalTasks", totalTasks,
		"concurrentLimit", concurrentLimit,
		"duration", duration)

	startTime := time.Now()
	deadline := startTime.Add(duration)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrentLimit)
	executionTimes := make([]time.Duration, 0, totalTasks)
	errors := make([]error, 0, totalTasks)

	taskCount := 0
	for time.Now().Before(deadline) && taskCount < totalTasks {
		select {
		case semaphore <- struct{}{}:
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-semaphore }()

				execTime, err := pt.executeSingleTask(ctx, task)
				executionTimes = append(executionTimes, execTime)
				errors = append(errors, err)
			}()
			taskCount++
		case <-ctx.Done():
			goto loopEnd
		}
	}
loopEnd:

	wg.Wait()
	totalTime := time.Since(startTime)
	pt.calculateMetrics(executionTimes, errors, totalTime, true)

	pt.logger.Info("Load test completed",
		"tasksExecuted", taskCount,
		"totalTime", totalTime,
		"throughput", pt.metrics.TaskThroughput)

	return pt.metrics
}

// StressTest pushes the system to its limits
func (pt *PerformanceTester) StressTest(
	ctx context.Context,
	task *task_engine.Task,
	initialConcurrency int,
	maxConcurrency int,
	stepDuration time.Duration,
) *PerformanceMetrics {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.logger.Info("Starting stress test",
		"initialConcurrency", initialConcurrency,
		"maxConcurrency", maxConcurrency,
		"stepDuration", stepDuration)

	var allExecutionTimes []time.Duration
	var allErrors []error
	var totalTime time.Duration

	for concurrency := initialConcurrency; concurrency <= maxConcurrency; concurrency *= 2 {
		pt.logger.Info("Testing concurrency level", "concurrency", concurrency)

		stepStart := time.Now()
		stepMetrics := pt.LoadTest(ctx, task, concurrency*10, concurrency, stepDuration)

		// Collect metrics from this step
		allExecutionTimes = append(allExecutionTimes, stepMetrics.AverageExecutionTime)
		allErrors = append(allErrors, nil) // Simplified for this example

		stepTime := time.Since(stepStart)
		totalTime += stepTime
		if stepMetrics.ErrorRate > 50 || stepMetrics.AverageExecutionTime > 10*time.Second {
			pt.logger.Warn("System showing signs of stress",
				"concurrency", concurrency,
				"errorRate", stepMetrics.ErrorRate,
				"avgTime", stepMetrics.AverageExecutionTime)
			break
		}
	}

	pt.calculateMetrics(allExecutionTimes, allErrors, totalTime, true)
	pt.logger.Info("Stress test completed", "totalTime", totalTime)

	return pt.metrics
}

// GetMetrics returns the current performance metrics
func (pt *PerformanceTester) GetMetrics() *PerformanceMetrics {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.metrics
}

// ResetMetrics resets all performance metrics
func (pt *PerformanceTester) ResetMetrics() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.metrics = &PerformanceMetrics{}
}

// GenerateReport generates a comprehensive performance report
func (pt *PerformanceTester) GenerateReport() map[string]interface{} {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	report := map[string]interface{}{
		"timestamp":              time.Now().Format(time.RFC3339),
		"total_tasks_executed":   pt.metrics.TotalTasksExecuted,
		"total_execution_time":   pt.metrics.TotalExecutionTime.String(),
		"average_execution_time": pt.metrics.AverageExecutionTime.String(),
		"min_execution_time":     pt.metrics.MinExecutionTime.String(),
		"max_execution_time":     pt.metrics.MaxExecutionTime.String(),
		"concurrent_tasks":       pt.metrics.ConcurrentTasks,
		"task_throughput":        pt.metrics.TaskThroughput,
		"error_rate":             pt.metrics.ErrorRate,
		"performance_score":      pt.calculatePerformanceScore(),
	}

	return report
}

// calculatePerformanceScore calculates a performance score based on metrics
func (pt *PerformanceTester) calculatePerformanceScore() float64 {
	if pt.metrics.TotalTasksExecuted == 0 {
		return 0
	}

	// Simple scoring algorithm - can be enhanced based on requirements
	throughputScore := pt.metrics.TaskThroughput / 100 // Normalize to 0-1
	errorPenalty := pt.metrics.ErrorRate / 100
	timingScore := 1.0 - (pt.metrics.AverageExecutionTime.Seconds() / 10.0) // Normalize to 0-1

	if timingScore < 0 {
		timingScore = 0
	}

	score := (throughputScore + timingScore) / 2 * (1 - errorPenalty)
	if score < 0 {
		score = 0
	}

	return score * 100 // Return as percentage
}
