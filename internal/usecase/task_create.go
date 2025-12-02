package usecase

import (
	"bcc-go-project/internal/domain/entity"
	_ "bcc-go-project/pkg/detach_context"
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
		return 0, "", fmt.Errorf("TaskService.CreateTask error: %w", err)
	}
	task.Id = idTask

	// обработка URLов
	go func() {
		wg := &sync.WaitGroup{}
		// создаем независимую копию контекста
		//detachCtx := dctx.DetachContext(ctx)
		// от него создаем контекст для загрузчиков с таймаутом такса
		loadCtx, cancel := context.WithTimeout(context.Background(), task.Timeout*time.Second)
		defer cancel()
		for _, file := range task.Files {
			//запускаем скачивание файлов асинхронно
			wg.Add(1)
			go func(ctx context.Context) {
				//defer func() {
				//	// сохраняем информацию по обработке URLа в репозиторий
				//	// не понятно как правильно обработать ошибку сохранения если оно асинхронное
				//	_, _ = ts.FileRepository.Create(ctx, file, task.Id)
				//	wg.Done()
				//}()
				defer wg.Done()
				if data, err := ts.HttpLoader.Load(loadCtx, file.Url); err != nil {
					file.Error = err
					log.Printf("Ошибка загрузки файла: %s", err)
				} else {
					file.Data = data
					log.Printf("Файл по ссылке %s загружен", file.Url)
				}
			}(loadCtx)
		}
		//ждем завершение загрузок или таймаута
		wg.Wait()
		err := ts.TaskRepository.UpdateStatus(loadCtx, task.Id, entity.TaskStatusDone)
		if err != nil {
			log.Printf("CreateTask.UpdateStatus error: %s", err)
		}

	}()

	// отправляем успех
	return task.Id, task.Status, nil
}
