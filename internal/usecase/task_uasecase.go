package usecase

import (
	"bcc-go-project/internal/domain/entity"
	"context"
)

type TaskRepository interface {
	Create(ctx context.Context, task entity.Task) (id entity.IdTask, err error)
	Get(ctx context.Context, id entity.IdTask) (*entity.Task, error)
	UpdateStatus(ctx context.Context, id entity.IdTask, status string) error
}
type FileRepository interface {
	Create(ctx context.Context, file entity.File, idTask entity.IdTask) (id entity.IdFile, err error)
	UpdateData(ctx context.Context, id entity.IdFile, data []byte) error
	UpdateErr(ctx context.Context, id entity.IdFile, err error) error
}

type HttpLoader interface {
	Load(ctx context.Context, url entity.Url) ([]byte, error)
}

type TaskUseCase struct {
	TaskRepository TaskRepository
	FileRepository FileRepository
	HttpLoader     HttpLoader
}

func NewTaskUseCase(taskRepo TaskRepository, fileRepo FileRepository, httpLoader HttpLoader) *TaskUseCase {
	return &TaskUseCase{
		TaskRepository: taskRepo,
		FileRepository: fileRepo,
		HttpLoader:     httpLoader,
	}
}
