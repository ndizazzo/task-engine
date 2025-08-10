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
}

// TaskInterface defines the contract for individual tasks
type TaskInterface interface {
	GetID() string
	GetName() string
	Run(ctx context.Context) error
	GetCompletedTasks() int
	GetTotalTime() time.Duration
}

// ResultProvider interface for tasks that produce results
type ResultProvider interface {
	GetResult() interface{}
	GetError() error
}
