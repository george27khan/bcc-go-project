package usecase

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
)

// GetTask получить таск
func (ts *TaskUseCase) GetTask(ctx context.Context, id entity.IdTask) (*entity.Task, error) {
	//получаем таск
	task, err := ts.TaskRepository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("TaskService.GetTask: %w", err)
	}
	// отправляем успех
	return task, nil
}
