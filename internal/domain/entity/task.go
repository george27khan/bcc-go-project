package entity

import "time"

type Task struct {
	Id      int
	Timeout time.Duration
	Status  string
	Created time.Time
}

func NewTask(timeout time.Duration) *Task {
	return &Task{
		Timeout: timeout,
		Status:  "PROCESS",
		Created: time.Now(),
	}
}
