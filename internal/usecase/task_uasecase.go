package usecase

import (
	"bcc-go-project/internal/domain/entity"
	"context"
)

type TaskRepository interface {
	Create(ctx context.Context, task entity.Task) (id entity.IdTask, err error)
	Get(ctx context.Context, id entity.IdTask) (*entity.Task, error)
	UpdateStatus(ctx context.Context, id entity.IdTask, status string) error
	UpdateFileData(ctx context.Context, id entity.IdTask, url entity.Url, data []byte) error
	UpdateFileErr(ctx context.Context, id entity.IdTask, url entity.Url, fileErr error) error
	GetTaskFile(ctx context.Context, idTask entity.IdTask, idFile entity.IdFile) ([]byte, error)
}

type HttpLoader interface {
	Load(ctx context.Context, url entity.Url) ([]byte, error)
}

type TaskUseCase struct {
	TaskRepository TaskRepository
	HttpLoader     HttpLoader
}

func NewTaskUseCase(taskRepo TaskRepository, httpLoader HttpLoader) *TaskUseCase {
	return &TaskUseCase{
		TaskRepository: taskRepo,
		HttpLoader:     httpLoader,
	}
}
