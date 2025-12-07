package usecase

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
	"time"
)

type GetTaskRepository interface {
	Get(ctx context.Context, id entity.IdTask) (*entity.Task, error)
}

type GetTaskUseCase struct {
	Repository GetTaskRepository
}

func NewGetTaskUseCase(getTaskRepo GetTaskRepository) *GetTaskUseCase {
	return &GetTaskUseCase{
		Repository: getTaskRepo,
	}
}

// GetTask получить таск
func (ts *GetTaskUseCase) GetTask(ctx context.Context, id entity.IdTask) (*entity.Task, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("TaskService.GetTask: %w", ctx.Err())
	}
	//получаем таск
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	task, err := ts.Repository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("TaskService.GetTask: %w", err)
	}
	// отправляем успех
	return task, nil
}
