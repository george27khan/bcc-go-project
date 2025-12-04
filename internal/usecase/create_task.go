package usecase

import (
	"bcc-go-project/internal/domain/entity"
	dctx "bcc-go-project/internal/pkg/detach_context"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// CreateTask функция создания таска
func (ts *TaskUseCase) CreateTask(ctx context.Context, task entity.Task) (idTask entity.IdTask, status string, err error) {

	//создаем таск в репо
	idTask, err = ts.TaskRepository.Create(ctx, task)
	if err != nil {
		return 0, "", fmt.Errorf("TaskService.CreateTask: %w", err)
	}
	task.Id = idTask

	// асинхронный запуск загрузки URLов
	go func() {
		wg := &sync.WaitGroup{}
		// создаем независимую копию контекста
		detachCtx := dctx.DetachContext(ctx)
		// от него создаем контекст для загрузчиков с общим таймаутом таска
		loadCtx, cancel := context.WithTimeout(detachCtx, task.Timeout*time.Second)
		defer cancel()
		for _, file := range task.Files {
			//запускаем скачивание файлов асинхронно
			wg.Add(1)
			go func() {
				defer wg.Done()
				if data, err := ts.HttpLoader.Load(loadCtx, file.Url); err != nil {
					file.Error = err
					_ = ts.TaskRepository.UpdateFileErr(loadCtx, idTask, file.Url, file.Error)
					log.Printf("Ошибка загрузки файла: %s", err)
				} else {
					file.Data = data
					_ = ts.TaskRepository.UpdateFileData(loadCtx, idTask, file.Url, file.Data)
					log.Printf("Файл по ссылке %s загружен", file.Url)
				}
			}()
		}
		//ждем завершение загрузок или таймаута
		wg.Wait()
		err := ts.TaskRepository.UpdateStatus(loadCtx, task.Id, entity.TaskStatusDone)
		if err != nil {
			log.Printf("CreateTask.UpdateStatus: %s", err)
		}
	}()

	// отправляем успех
	return task.Id, task.Status, nil
}
