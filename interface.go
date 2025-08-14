package task_engine

import (
	"context"
	"time"
)

// TaskManagerInterface defines the contract for task management
type TaskManagerInterface interface {
	AddTask(task *Task) error
	RunTask(taskID string) error
	StopTask(taskID string) error
	StopAllTasks()
	GetRunningTasks() []string
	IsTaskRunning(taskID string) bool
	GetGlobalContext() *GlobalContext
	ResetGlobalContext()
}

// TaskInterface defines the contract for individual tasks
type TaskInterface interface {
	GetID() string
	GetName() string
	Run(ctx context.Context) error
	RunWithContext(ctx context.Context, globalContext *GlobalContext) error
	GetCompletedTasks() int
	GetTotalTime() time.Duration
}

// ResultProvider interface for tasks that produce results
type ResultProvider interface {
	GetResult() interface{}
	GetError() error
}

// TaskWithResults interface for tasks that can optionally provide rich results
// Combines the task lifecycle with the ability to provide results and errors
// after execution, mirroring ActionWithResults.
type TaskWithResults interface {
	TaskInterface
	ResultProvider
}
