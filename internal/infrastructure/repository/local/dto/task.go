package dto

import (
	"bcc-go-project/internal/domain/entity"
	"time"
)

type Task struct {
	Id      entity.IdTask
	Timeout time.Duration
	Status  string
}

func ToTaskDomain(task Task) *entity.Task {
	return &entity.Task{
		Id:      task.Id,
		Timeout: task.Timeout,
		Status:  task.Status,
	}
}

func ToTaskRepo(task entity.Task) *Task {
	return &Task{
		Id:      task.Id,
		Timeout: task.Timeout,
		Status:  task.Status,
	}
}
