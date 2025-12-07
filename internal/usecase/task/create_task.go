package task

import (
	"bcc-go-project/internal/domain/entity"
	dctx "bcc-go-project/internal/pkg/detach_context"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

//go:generate mockgen -package task -source=create_task.go -destination=mock_create_task.go
type CreateTaskRepository interface {
	Create(ctx context.Context, task entity.Task) (id entity.IdTask, err error)
	UpdateStatus(ctx context.Context, id entity.IdTask, status entity.Status) error
	UpdateFileData(ctx context.Context, id entity.IdTask, url entity.Url, data []byte) error
	UpdateFileErr(ctx context.Context, id entity.IdTask, url entity.Url, fileErr error) error
}

type HttpLoader interface {
	Load(ctx context.Context, url entity.Url) ([]byte, error)
}

type CreateTaskUseCase struct {
	Repository CreateTaskRepository
	HttpLoader HttpLoader
}

func NewCreateTaskUseCase(createTaskRepo CreateTaskRepository, httpLoader HttpLoader) *CreateTaskUseCase {
	return &CreateTaskUseCase{
		Repository: createTaskRepo,
		HttpLoader: httpLoader,
	}
}

// CreateTask функция создания таска
func (ts *CreateTaskUseCase) CreateTask(ctx context.Context, task entity.Task) (idTask entity.IdTask, status entity.Status, err error) {
	if ctx.Err() != nil {
		return 0, "", fmt.Errorf("TaskService.CreateTask: %w", ctx.Err())
	}
	//создаем таск в репо
	idTask, err = ts.Repository.Create(ctx, task)
	if err != nil {
		return 0, "", fmt.Errorf("TaskService.CreateTask: %w", err)
	}

	task.Id = idTask

	// асинхронный запуск загрузки URLов
	go func() {
		wg := &sync.WaitGroup{}
		detachCtx := dctx.DetachContext(ctx)                                        // создаем независимую копию контекста т.к основной протухнет при ответе
		loadCtx, cancel := context.WithTimeout(detachCtx, task.Timeout*time.Second) // от него создаем контекст для загрузчиков с общим таймаутом таска
		defer cancel()
		for _, file := range task.Files {
			//запускаем скачивание файлов асинхронно
			wg.Add(1)
			go func() {
				defer wg.Done()
				if data, err := ts.HttpLoader.Load(loadCtx, file.Url); err != nil {
					file.Error = err
					ctxRep, cancelRep := context.WithTimeout(loadCtx, 100*time.Millisecond)
					defer cancelRep()
					_ = ts.Repository.UpdateFileErr(ctxRep, idTask, file.Url, file.Error)
					log.Printf("Ошибка загрузки файла taskId=%v; url=%s : %s", idTask, file.Url, err)
				} else {
					file.Data = data
					ctxRep, cancelRep := context.WithTimeout(loadCtx, 100*time.Millisecond)
					defer cancelRep()
					_ = ts.Repository.UpdateFileData(ctxRep, idTask, file.Url, file.Data)
					log.Printf("Файл загружен taskId=%v; url=%s", idTask, file.Url)
				}
			}()
		}
		//ждем завершение загрузок или таймаута
		wg.Wait()
		ctxRep, cancelRep := context.WithTimeout(loadCtx, 100*time.Millisecond)
		defer cancelRep()
		err := ts.Repository.UpdateStatus(ctxRep, task.Id, entity.TaskStatusDone)
		if err != nil {
			log.Printf("CreateTask.UpdateStatus taskId=%v: %s", idTask, err)
		} else {
			log.Printf("Загрузка таска завершена taskId=%v", idTask)
		}
	}()

	// отправляем ответ
	return task.Id, task.Status, nil
}
