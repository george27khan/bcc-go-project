package local

import (
	"bcc-go-project/internal/domain/entity"
	"bcc-go-project/internal/infrastructure/repository/local/dto"
	"bcc-go-project/internal/usecase"
	"context"
	"sync"
)

var _ usecase.TaskRepository = (*TaskRepository)(nil)

type TaskRepository struct {
	tasks []dto.Task
	mu    *sync.RWMutex
	id    entity.IdTask
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: make([]dto.Task, 0),
		mu:    &sync.RWMutex{},
	}
}

// Create создание таска в локальном хранилище
func (r *TaskRepository) Create(ctx context.Context, task entity.Task) (id entity.IdTask, err error) {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	taskRepo := dto.ToTaskRepo(task)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks = append(r.tasks, *taskRepo) // пишем новую таску
	taskRepo.Id = r.id                   //выдем таску ID
	r.id++                               // сдвигаем счетчик
	return taskRepo.Id, nil
}

// UpdateStatus обновление статуса таска
func (r *TaskRepository) UpdateStatus(ctx context.Context, id entity.IdTask, status string) error {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[id].Status = status // пишем новую таску
	return nil
}

// Get получение таска
func (r *TaskRepository) Get(ctx context.Context, id entity.IdTask) (*entity.Task, error) {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	taskRepo := r.tasks[id] // пишем новую таску
	task := dto.ToTaskDomain(taskRepo)
	return task, nil
}
