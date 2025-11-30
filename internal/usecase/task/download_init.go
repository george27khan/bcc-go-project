package task

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
)

type TaskRepositiry interface {
	Create(ctx context.Context, task entity.Task) (id int, err error)
}
type FileRepositiry interface {
	Create(ctx context.Context, task entity.File) (id int, err error)
}

type HttpLoader interface {
	Load(url string) ([]byte, error)
}

type DownloadInitUseCase struct {
	TaskRepositiry TaskRepositiry
	FileRepositiry FileRepositiry
	HttpLoader     HttpLoader
}

func NewTaskService(taskRepo TaskRepositiry, fileRepo FileRepositiry, httpLoader HttpLoader) *DownloadInitUseCase {

	return &DownloadInitUseCase{
		TaskRepositiry: taskRepo,
		FileRepositiry: fileRepo,
		HttpLoader:     httpLoader,
	}
}

// Create функция создания таска
func (ts DownloadInitUseCase) CreateTask(ctx context.Context, task entity.Task, files []entity.File) (id int, err error) {
	id, err = ts.TaskRepositiry.Create(ctx, task) //создаем таск
	if err != nil {
		return 0, fmt.Errorf("TaskService.Create error: %w", err)
	}

	// привязываем файлы к таску
	for _, file := range files {
		file.LoaderId = id
	}

	//запускаем скачивание файлов

	// сохраняем скаченные файлы в репозиторий

	// отправляем успех

	return id, nil
}
