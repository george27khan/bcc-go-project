package local

import (
	"bcc-go-project/internal/domain/entity"
	"bcc-go-project/internal/infrastructure/repository/errors"
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
		tasks: make(map[entity.IdTask]entity.Task, 100),
		mu:    &sync.RWMutex{},
	}
}

// Create создание таска в локальном хранилище
func (r *TaskRepository) Create(ctx context.Context, task entity.Task) (id entity.IdTask, err error) {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("TaskRepository.Create: %w", ctx.Err())
	default:
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
	select {
	case <-ctx.Done():
		return fmt.Errorf("TaskRepository.UpdateStatus: %w", ctx.Err())
	default:
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if task, ok := r.tasks[id]; !ok {
		return fmt.Errorf("TaskRepository.UpdateStatus: %w", errors.ErrTaskNotExist)
	} else {
		task.Status = status
		r.tasks[id] = task // обновляем таску
	}
	return nil
}

// Get получение таска
func (r *TaskRepository) Get(ctx context.Context, id entity.IdTask) (*entity.Task, error) {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("TaskRepository.Get: %w", ctx.Err())
	default:
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if task, ok := r.tasks[id]; !ok {
		return nil, fmt.Errorf("TaskRepository.Get: %w", errors.ErrTaskNotExist)
	} else {
		return &task, nil
	}
}

// UpdateFileData обновление данных
func (r *TaskRepository) UpdateFileData(ctx context.Context, id entity.IdTask, url entity.Url, data []byte) error {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return fmt.Errorf("TaskRepository.UpdateFileData: %w", ctx.Err())
	default:
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if task, ok := r.tasks[id]; !ok {
		return fmt.Errorf("TaskRepository.UpdateFileData: %w", errors.ErrTaskNotExist)
	} else {
		for i, file := range task.Files {
			if file.Url == url { // допускаем что url в рамках таска уникальные
				task.Files[i].Data = data
				task.Files[i].Id = r.idFile
				r.idFile++
				r.tasks[id] = task
				return nil
			}
		}
	}
	return fmt.Errorf("TaskRepository.UpdateFileData: %w", errors.ErrFileNotExist)
}

// UpdateFileErr обновление ошибки
func (r *TaskRepository) UpdateFileErr(ctx context.Context, id entity.IdTask, url entity.Url, fileErr error) error {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return fmt.Errorf("TaskRepository.UpdateFileErr: %w", ctx.Err())
	default:
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if task, ok := r.tasks[id]; !ok {
		return fmt.Errorf("TaskRepository.UpdateFileErr: %w", errors.ErrTaskNotExist)
	} else {
		for i, file := range task.Files {
			if file.Url == url { // допускаем что url в рамках таска уникальные
				task.Files[i].Error = fileErr
				r.tasks[id] = task
				return nil
			}
		}
	}
	return fmt.Errorf("TaskRepository.UpdateFileErr: %w", errors.ErrFileNotExist)
}

// GetTaskFile получение файла
func (r *TaskRepository) GetTaskFile(ctx context.Context, idTask entity.IdTask, idFile entity.IdFile) ([]byte, error) {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return []byte{}, fmt.Errorf("TaskRepository.GetTaskFile: %w", ctx.Err())
	default:
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if task, ok := r.tasks[idTask]; !ok {
		return []byte{}, fmt.Errorf("TaskRepository.GetTaskFile: %w", errors.ErrTaskNotExist)
	} else {
		for _, file := range task.Files {
			if file.Id == idFile {
			}
			return file.Data, nil
		}
	}
	return []byte{}, fmt.Errorf("TaskRepository.GetTaskFile: %w", errors.ErrFileNotExist)
}
