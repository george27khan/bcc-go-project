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

type BackgroundRunner interface {
	GoTask(ctx context.Context, task entity.Task, f func(context.Context, entity.Task))
	GoFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File, f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File))
}

type AsyncRunner struct{}

func (ar *AsyncRunner) GoTask(ctx context.Context, task entity.Task, f func(context.Context, entity.Task)) {
	go f(ctx, task)
}

func (ar *AsyncRunner) GoFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File, f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File)) {
	go f(ctx, wg, idTask, file)
}

type SyncRunner struct{}

func (sr *SyncRunner) GoTask(ctx context.Context, object entity.Task, f func(context.Context, entity.Task)) {
	f(ctx, object)
}
func (ar *SyncRunner) GoFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File, f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File)) {
	f(ctx, wg, idTask, file)
}

type HttpLoader interface {
	Load(ctx context.Context, url entity.Url) ([]byte, error)
}

type CreateTaskUseCase struct {
	Repository CreateTaskRepository
	HttpLoader HttpLoader
	Runner     BackgroundRunner
}

func NewCreateTaskUseCase(createTaskRepo CreateTaskRepository,
	httpLoader HttpLoader,
	runner BackgroundRunner,
) *CreateTaskUseCase {
	return &CreateTaskUseCase{
		Repository: createTaskRepo,
		HttpLoader: httpLoader,
		Runner:     runner,
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
	ts.Runner.GoTask(ctx, task, ts.RunDownload)
	// отправляем ответ
	return task.Id, task.Status, nil
}

// RunDownload запуск загрузок
func (ts *CreateTaskUseCase) RunDownload(ctx context.Context, task entity.Task) {
	wg := &sync.WaitGroup{}
	detachCtx := dctx.DetachContext(ctx)                                        // создаем независимую копию контекста т.к основной протухнет при ответе
	loadCtx, cancel := context.WithTimeout(detachCtx, task.Timeout*time.Second) // от него создаем контекст для загрузчиков с общим таймаутом таска
	defer cancel()
	for _, file := range task.Files {
		//запускаем скачивание файлов асинхронно
		wg.Add(1)
		ts.Runner.GoFile(loadCtx, wg, task.Id, file, ts.DownloadFile)
	}
	//ждем завершение загрузок или таймаута
	wg.Wait()
	ctxRep, cancelRep := context.WithTimeout(loadCtx, 100*time.Millisecond)
	defer cancelRep()
	err := ts.Repository.UpdateStatus(ctxRep, task.Id, entity.TaskStatusDone)
	if err != nil {
		log.Printf("CreateTask.UpdateStatus taskId=%v: %s", task.Id, err)
	} else {
		log.Printf("Загрузка таска завершена taskId=%v", task.Id)
	}
}

// DownloadFile запуск скачивания файла
func (ts *CreateTaskUseCase) DownloadFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File) {
	defer wg.Done()
	if data, err := ts.HttpLoader.Load(ctx, file.Url); err != nil {
		file.Error = err
		ctxRep, cancelRep := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancelRep()
		_ = ts.Repository.UpdateFileErr(ctxRep, idTask, file.Url, file.Error)
		log.Printf("Ошибка загрузки файла taskId=%v; url=%s : %s", idTask, file.Url, err)
	} else {
		file.Data = data
		ctxRep, cancelRep := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancelRep()
		_ = ts.Repository.UpdateFileData(ctxRep, idTask, file.Url, file.Data)
		log.Printf("Файл загружен taskId=%v; url=%s", idTask, file.Url)
	}
}
