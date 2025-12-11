package local

import (
	"bcc-go-project/internal/domain/entity"
	"bcc-go-project/internal/infrastructure/repository/errors_repo"
	"bcc-go-project/internal/usecase/task"
	"context"
	"fmt"
	"sync"
)

var (
	_ task.CreateTaskRepository = (*TaskRepository)(nil)
	_ task.GetTaskRepository    = (*TaskRepository)(nil)
	_ task.TaskFileRepository   = (*TaskRepository)(nil)
)

type TaskRepository struct {
	tasks  map[entity.IdTask]entity.Task
	mu     *sync.RWMutex
	idTask entity.IdTask
	idFile entity.IdFile
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks:  make(map[entity.IdTask]entity.Task, 100),
		mu:     &sync.RWMutex{},
		idFile: entity.IdFile(1),
		idTask: entity.IdTask(1),
	}
}

// Create создание таска в локальном хранилище
func (r *TaskRepository) Create(ctx context.Context, task entity.Task) (id entity.IdTask, err error) {
	// завершение операции по контексту
	if ctx.Err() != nil {
		return 0, fmt.Errorf("TaskRepository.Create: %w", ctx.Err())
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	task.Id = r.idTask
	r.tasks[r.idTask] = task // пишем новую таску
	r.idTask++               // сдвигаем счетчик
	return task.Id, nil
}

// UpdateStatus обновление статуса таска
func (r *TaskRepository) UpdateStatus(ctx context.Context, id entity.IdTask, status entity.Status) error {
	// завершение операции по контексту
	if ctx.Err() != nil {
		return fmt.Errorf("TaskRepository.UpdateStatus: %w", ctx.Err())
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if tsk, ok := r.tasks[id]; !ok {
		return fmt.Errorf("TaskRepository.UpdateStatus: %w", errors_repo.ErrTaskNotExist)
	} else {
		tsk.Status = status
		r.tasks[id] = tsk // обновляем таску
	}
	return nil
}

// Get получение таска
func (r *TaskRepository) Get(ctx context.Context, id entity.IdTask) (*entity.Task, error) {
	// завершение операции по контексту
	if ctx.Err() != nil {
		return nil, fmt.Errorf("TaskRepository.Get: %w", ctx.Err())
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if tsk, ok := r.tasks[id]; !ok {
		return nil, fmt.Errorf("TaskRepository.Get: %w", errors_repo.ErrTaskNotExist)
	} else {
		return &tsk, nil
	}
}

// UpdateFileData обновление данных
func (r *TaskRepository) UpdateFileData(ctx context.Context, id entity.IdTask, url entity.Url, data []byte) error {
	// завершение операции по контексту
	if ctx.Err() != nil {
		return fmt.Errorf("TaskRepository.UpdateFileData: %w", ctx.Err())
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if tsk, ok := r.tasks[id]; !ok {
		return fmt.Errorf("TaskRepository.UpdateFileData: %w", errors_repo.ErrTaskNotExist)
	} else {
		for i, file := range tsk.Files {
			if file.Url == url { // допускаем что url в рамках таска уникальные
				tsk.Files[i].Data = data
				tsk.Files[i].Id = r.idFile
				r.idFile++
				r.tasks[id] = tsk
				return nil
			}
		}
	}
	return fmt.Errorf("TaskRepository.UpdateFileData: %w", errors_repo.ErrFileNotExist)
}

// UpdateFileErr обновление ошибки
func (r *TaskRepository) UpdateFileErr(ctx context.Context, id entity.IdTask, url entity.Url, fileErr error) error {
	// завершение операции по контексту
	if ctx.Err() != nil {
		return fmt.Errorf("TaskRepository.UpdateFileErr: %w", ctx.Err())
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if tsk, ok := r.tasks[id]; !ok {
		return fmt.Errorf("TaskRepository.UpdateFileErr: %w", errors_repo.ErrTaskNotExist)
	} else {
		for i, file := range tsk.Files {
			if file.Url == url { // допускаем что url в рамках таска уникальные
				tsk.Files[i].Error = fileErr
				r.tasks[id] = tsk
				return nil
			}
		}
	}
	return fmt.Errorf("TaskRepository.UpdateFileErr: %w", errors_repo.ErrFileNotExist)
}

// GetTaskFile получение файла
func (r *TaskRepository) GetTaskFile(ctx context.Context, idTask entity.IdTask, idFile entity.IdFile) ([]byte, error) {
	// завершение операции по контексту
	if ctx.Err() != nil {
		return nil, fmt.Errorf("TaskRepository.GetTaskFile: %w", ctx.Err())
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if tsk, ok := r.tasks[idTask]; !ok {
		return []byte{}, fmt.Errorf("TaskRepository.GetTaskFile: %w", errors_repo.ErrTaskNotExist)
	} else {
		for _, file := range tsk.Files {
			if file.Id == idFile {
				return file.Data, nil
			}
		}
	}
	return []byte{}, fmt.Errorf("TaskRepository.GetTaskFile: %w", errors_repo.ErrFileNotExist)
}
