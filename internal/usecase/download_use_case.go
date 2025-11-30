package usecase

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
)

type TaskRepositiry interface {
	Create(ctx context.Context, task entity.Task) (id int, err error)
}

type DownloadUseCase struct {
	TaskRepositiry TaskRepositiry
}

func NewTaskService(TaskRepositiry TaskRepositiry) *TaskUseCase {

	return &TaskUseCase{TaskRepositiry: TaskRepositiry}
}

// Create функция создания таска
func (ts TaskUseCase) Create(ctx context.Context, task entity.Task) (id int, err error) {
	id, err = ts.TaskRepositiry.Create(ctx, task)
	if err != nil {
		return 0, fmt.Errorf("TaskService.Create error: %w", err)
	}
	return id, nil
}
