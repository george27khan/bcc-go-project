package task

import (
	"bcc-go-project/internal/domain/entity"
	"context"
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
	UpdateFileErr(ctx context.Context, id entity.IdTask, url entity.Url, fileErr error) error
}

type BackgroundRunner interface {
	GoTask(ctx context.Context, task entity.Task, f func(context.Context, *sync.WaitGroup, entity.Task))
	GoFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File, f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File))
}

// AsyncRunner реализация асинхронного запуска загрузки
type AsyncRunner struct {
	WgRoot *sync.WaitGroup
}

func (ar *AsyncRunner) GoTask(ctx context.Context, task entity.Task, f func(context.Context, *sync.WaitGroup, entity.Task)) {
	go f(ctx, ar.WgRoot, task)
}
func (ar *AsyncRunner) GoFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File, f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File)) {
	go f(ctx, wg, idTask, file)
}

type HttpLoader interface {
	Load(ctx context.Context, url entity.Url) ([]byte, error)
}

type CreateTaskUseCase struct {
	Repository CreateTaskRepository
	HttpLoader HttpLoader
	Runner     BackgroundRunner
	RootCtx    context.Context
}

func NewCreateTaskUseCase(createTaskRepo CreateTaskRepository,
	httpLoader HttpLoader,
	runner BackgroundRunner,
	rootCtx context.Context,
) *CreateTaskUseCase {
	return &CreateTaskUseCase{
		Repository: createTaskRepo,
		HttpLoader: httpLoader,
		Runner:     runner,
		RootCtx:    rootCtx,
	}
}

// CreateTask функция создания таска
func (ts *CreateTaskUseCase) CreateTask(ctx context.Context, task entity.Task) (idTask entity.IdTask, status entity.Status, err error) {
	if ctx.Err() != nil {
		return 0, "", fmt.Errorf("TaskService.CreateTask: %w", ctx.Err())
	}

	idTask, err = ts.Repository.Create(ctx, task) //создаем таск в репо
	if err != nil {
		return 0, "", fmt.Errorf("TaskService.CreateTask: %w", err)
	}

	task.Id = idTask
	ts.Runner.GoTask(ts.RootCtx, task, ts.RunDownload) // передаем корневой контекст
	return task.Id, task.Status, nil
}

// RunDownload запуск загрузок
func (ts *CreateTaskUseCase) RunDownload(ctx context.Context, wgRoot *sync.WaitGroup, task entity.Task) {
	wgRoot.Add(1)
	defer wgRoot.Done()
	wg := &sync.WaitGroup{}
	//detachCtx := dctx.DetachContext(ctx)                                        // создаем независимую копию контекста т.к основной протухнет при ответе
	loadCtx, cancel := context.WithTimeout(ctx, task.Timeout*time.Second) // от него создаем контекст для загрузчиков с общим таймаутом таска
	defer cancel()
	for _, file := range task.Files {
		if ctx.Err() != nil {
			log.Printf("RunDownload завершен по контексту: %v", ctx.Err())
			return
		}
		//запускаем скачивание файлов асинхронно
		wg.Add(1)
		ts.Runner.GoFile(loadCtx, wg, task.Id, file, ts.DownloadFile)
	}
	//ждем завершение загрузок или таймаута
	wg.Wait()
	ctxRep, cancelRep := context.WithTimeout(loadCtx, repCtxTimeout)
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
	if ctx.Err() != nil {
		log.Printf("DownloadFile завершен по контексту: %v", ctx.Err())
		return
	}
	//time.Sleep(60 * time.Second)
	if data, err := ts.HttpLoader.Load(ctx, file.Url); err != nil {
		file.Error = err
		ctxRep, cancelRep := context.WithTimeout(ctx, repCtxTimeout)
		defer cancelRep()
		_ = ts.Repository.UpdateFileErr(ctxRep, idTask, file.Url, file.Error)
		log.Printf("Ошибка загрузки файла taskId=%v; url=%s : %s", idTask, file.Url, err)
	} else {
		file.Data = data
		ctxRep, cancelRep := context.WithTimeout(ctx, repCtxTimeout)
		defer cancelRep()
		_ = ts.Repository.UpdateFileData(ctxRep, idTask, file.Url, file.Data)
		log.Printf("Файл загружен taskId=%v; url=%s", idTask, file.Url)
	}
}
