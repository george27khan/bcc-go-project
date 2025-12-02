package task_usecase

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type TaskRepository interface {
	Create(ctx context.Context, task entity.Task) (id int, err error)
	UpdateStatus(ctx context.Context, id int, status string) error
}
type FileRepository interface {
	Create(ctx context.Context, file entity.File) (id int, err error)
	UpdateData(ctx context.Context, id int, data []byte) error
	UpdateErr(ctx context.Context, id int, err error) error
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
func (ts DownloadInitUseCase) CreateTask(ctx context.Context, task entity.Task, files []entity.File) (idTask int, status string, err error) {
	wg := &sync.WaitGroup{}
	//создаем таск
	idTask, err = ts.TaskRepository.Create(ctx, task)
	if err != nil {
		return 0, "", fmt.Errorf("TaskService.CreateTask error: %w", err)
	}

	for _, file := range files {
		file.LoaderId = idTask
		//запускаем скачивание файлов
		wg.Add(1)
		go func() {
			defer wg.Done()
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

	//запускаем процесс следящий за загрузкой для обновления статуса такса
	go func() {
		wg.Wait()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = ts.TaskRepository.UpdateStatus(ctx, task.Id, entity.TaskStatusDone)
	}()

	// отправляем успех

	return idTask, task.Status, nil
}
