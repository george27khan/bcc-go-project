package entity

import "time"

const (
	TaskStatusProcess = "PROCESS"
	TaskStatusDone    = "DONE"
)

type Task struct {
	Id      int
	Timeout time.Duration
	Status  string
	Created time.Time
}

func NewTask(timeout time.Duration) Task {
	return Task{
		Timeout: timeout,
		Status:  TaskStatusProcess,
	}
}
