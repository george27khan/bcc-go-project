package usecase

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
)

// GetTaskFile получить файл из таска
func (ts *TaskUseCase) GetTaskFile(ctx context.Context, idTask entity.IdTask, idFile entity.IdFile) ([]byte, error) {
	//получаем таск
	data, err := ts.TaskRepository.GetTaskFile(ctx, idTask, idFile)
	if err != nil {
		return nil, fmt.Errorf("TaskService.GetTaskFile: %w", err)
	}
	// отправляем успех
	return data, nil
}
