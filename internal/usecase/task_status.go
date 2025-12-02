package usecase

//
//import (
//	"bcc-go-project/internal/domain/entity"
//	dctx "bcc-go-project/pkg/detach_context"
//	"context"
//	"fmt"
//	"log"
//	"sync"
//	"time"
//)
//
//// GetTaskStatus функция получения состояния таска
//func (ts *TaskUseCase) GetTaskStatus(ctx context.Context, id entity.IdTask) (task *entity.Task, status string, err error) {
//
//	//создаем таск в репо
//	task, err = ts.TaskRepository.Get(ctx, id)
//	if err != nil {
//		return nil, "", fmt.Errorf("TaskService.CreateTask error: %w", err)
//	}
//
//		for _, file := range files {
//			file.LoaderId = idTask // привязываем файл к таску
//			//запускаем скачивание файлов асинхронно
//			wg.Add(1)
//			go func() {
//				defer func() {
//					// сохраняем информацию по обработке URLа в репозиторий
//					// не понятно как правильно обработать ошибку сохранения если оно асинхронное
//					_, _ = ts.FileRepository.Create(ctx, file)
//					wg.Done()
//				}()
//				if data, err := ts.HttpLoader.Load(loadCtx, file.Url); err != nil {
//					file.Error = err
//					log.Printf("Ошибка загрузки файла: %s", err)
//				} else {
//					file.Data = data
//					log.Printf("Файл по ссылке %s загружен", file.Url)
//				}
//			}()
//		}
//		//ждем завершение загрузок или таймаута
//		wg.Wait()
//		err := ts.TaskRepository.UpdateStatus(ctx, task.Id, entity.TaskStatusDone)
//		log.Printf("CreateTask.UpdateStatus error: %s", err)
//	}()
//
//	// отправляем успех
//	return idTask, task.Status, nil
//}
