package task

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
	"time"
)

//go:generate mockgen -package task -source=get_task_file.go -destination=mock_get_task_file.go
type TaskFileRepository interface {
	GetTaskFile(ctx context.Context, idTask entity.IdTask, idFile entity.IdFile) ([]byte, error)
}

type TaskFileUseCase struct {
	Repository TaskFileRepository
}

func NewTaskFileUseCase(taskFileRepo TaskFileRepository) *TaskFileUseCase {
	return &TaskFileUseCase{
		Repository: taskFileRepo,
	}
}

// GetTaskFile получить файл из таска
func (ts *TaskFileUseCase) GetTaskFile(ctx context.Context, idTask entity.IdTask, idFile entity.IdFile) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("TaskService.GetTaskFile: %w", ctx.Err())
	}
	//получаем таск
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	data, err := ts.Repository.GetTaskFile(ctx, idTask, idFile)
	if err != nil {
		return nil, fmt.Errorf("TaskService.GetTaskFile: %w", err)
	}
	// отправляем успех
	return data, nil
}
