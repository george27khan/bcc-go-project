package task

import (
	"bcc-go-project/internal/domain/entity"
	dctx "bcc-go-project/internal/pkg/detach_context"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const repCtxTimeout = 100 * time.Millisecond // Время жизни контекста для запроса в бд, в идеале вынести в настройки

//go:generate mockgen -package task -source=create_task.go -destination=mock_create_task.go
type CreateTaskRepository interface {
	Create(ctx context.Context, task entity.Task) (id entity.IdTask, err error)
	UpdateStatus(ctx context.Context, id entity.IdTask, status entity.Status) error
	UpdateFileData(ctx context.Context, id entity.IdTask, url entity.Url, data []byte) error
	UpdateFileErr(ctx context.Context, id entity.IdTask, url entity.Url, fileErr entity.Error) error
}

type BackgroundRunner interface {
	GoTask(ctx context.Context, task entity.Task, f func(context.Context, *sync.WaitGroup, entity.Task) error)
	GoFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File, f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File) error)
}

// AsyncRunner реализация асинхронного запуска загрузки
type AsyncRunner struct {
	WgRoot *sync.WaitGroup
}

func NewAsyncRunner() AsyncRunner {
	return AsyncRunner{WgRoot: &sync.WaitGroup{}}
}

func (ar AsyncRunner) GoTask(ctx context.Context, task entity.Task, f func(context.Context, *sync.WaitGroup, entity.Task) error) {
	go f(ctx, ar.WgRoot, task)
}
func (ar AsyncRunner) GoFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File, f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File) error) {
	go f(ctx, wg, idTask, file)
}

// SyncRunner реализация синхронного запуска загрузки для тестов
type SyncRunner struct{}

func (sr *SyncRunner) GoTask(ctx context.Context, object entity.Task, f func(context.Context, *sync.WaitGroup, entity.Task) error) {
	f(ctx, &sync.WaitGroup{}, object)
}
func (sr *SyncRunner) GoFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File,
	f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File) error) {
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
	idTask, err = ts.Repository.Create(ctx, task) //создаем таск в репо
	if err != nil {
		return 0, "", fmt.Errorf("TaskService.CreateTask: %w", err)
	}

	task.Id = idTask
	ts.Runner.GoTask(ctx, task, ts.RunDownload) // передаем корневой контекст
	return task.Id, task.Status, nil
}

// RunDownload запуск загрузок
func (ts *CreateTaskUseCase) RunDownload(ctx context.Context, wgRoot *sync.WaitGroup, task entity.Task) error {
	wgRoot.Add(1)
	defer wgRoot.Done()
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
	ctxRep, cancelRep := context.WithTimeout(loadCtx, repCtxTimeout)
	defer cancelRep()

	if err := ts.Repository.UpdateStatus(ctxRep, task.Id, entity.TaskStatusDone); err != nil {
		log.Printf("CreateTask.UpdateStatus taskId=%v: %s", task.Id, err)
		return err
	}

	log.Printf("Загрузка таска завершена taskId=%v", task.Id)
	return nil

}

// DownloadFile запуск скачивания файла
func (ts *CreateTaskUseCase) DownloadFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File) error {
	defer wg.Done()
	time.Sleep(20 * time.Second)
	data, err := ts.HttpLoader.Load(ctx, file.Url)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			file.Error = entity.FileErrTimeout
		} else {
			file.Error = entity.FileErr
		}
		log.Printf("Ошибка загрузки файла taskId=%v; url=%s : %s", idTask, file.Url, err)
		detachCtx := dctx.DetachContext(ctx) // создаем независимую копию контекста т.к основной может протухнуть и не успеем записать в бд
		ctxRep, cancelRep := context.WithTimeout(detachCtx, repCtxTimeout)
		defer cancelRep()
		if err = ts.Repository.UpdateFileErr(ctxRep, idTask, file.Url, file.Error); err != nil {
			return err
		}
		return err
	}
	log.Printf("Файл загружен taskId=%v; url=%s", idTask, file.Url)
	file.Data = data
	detachCtx := dctx.DetachContext(ctx) // создаем независимую копию контекста т.к основной может протухнуть и не успеем записать в бд
	ctxRep, cancelRep := context.WithTimeout(detachCtx, repCtxTimeout)
	defer cancelRep()
	if err = ts.Repository.UpdateFileData(ctxRep, idTask, file.Url, file.Data); err != nil {
		return err
	}
	return nil
}
