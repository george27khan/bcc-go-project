package task

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
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
	//получаем таск
	ctx, cancel := context.WithTimeout(ctx, repCtxTimeout)
	defer cancel()
	data, err := ts.Repository.GetTaskFile(ctx, idTask, idFile)
	if err != nil {
		return nil, fmt.Errorf("TaskService.GetTaskFile: %w", err)
	}
	// отправляем успех
	return data, nil
}
