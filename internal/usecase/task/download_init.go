package task

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
	"log"
	"time"
)

type TaskRepository interface {
	Create(ctx context.Context, task entity.Task) (id int, err error)
}
type FileRepository interface {
	Create(ctx context.Context, task entity.File) (id int, err error)
}

type HttpLoader interface {
	Load(ctx context.Context, url string) ([]byte, error)
}

type DownloadInitUseCase struct {
	TaskRepository TaskRepository
	FileRepository FileRepository
	HttpLoader     HttpLoader
}

func NewDownloadInitUseCase(taskRepo TaskRepository, fileRepo FileRepository, httpLoader HttpLoader) *DownloadInitUseCase {
	return &DownloadInitUseCase{
		TaskRepository: taskRepo,
		FileRepository: fileRepo,
		HttpLoader:     httpLoader,
	}
}

// CreateTask функция создания таска
func (ts DownloadInitUseCase) CreateTask(ctx context.Context, task entity.Task, files []entity.File) (id int, status string, err error) {
	id, err = ts.TaskRepository.Create(ctx, task) //создаем таск
	if err != nil {
		return 0, "", fmt.Errorf("TaskService.Create error: %w", err)
	}

	// привязываем файлы к таску
	for _, file := range files {
		file.LoaderId = id
		//запускаем скачивание файлов
		go func() {
			//таймаут
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //тут нужно сделать детач контекст
			defer cancel()
			if data, err := ts.HttpLoader.Load(ctx, file.Url); err != nil {
				file.Error = err
				log.Printf("Ошибка загрузки файла: %s", err)
			} else {
				file.Data = data
				log.Printf("Файл по ссылке %s загружен", file.Url)
				// сохраняем скаченные файлы в репозиторий
				// не понятно как правильно обработать ошибку сохранения если оно асинхронное
				_, _ = ts.FileRepository.Create(ctx, file)

			}
		}()
	}

	// отправляем успех

	return id, task.Status, nil
}
