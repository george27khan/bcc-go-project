package task

import (
	"bcc-go-project/internal/domain/entity"
	"bcc-go-project/internal/domain/service/task"
	"context"
	"sync"
)

var _ task.TaskRepositiry = (*TaskRepository)(nil)

type TaskRepository struct {
	tasks []entity.Task
	mu    *sync.Mutex
	id    int
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: make([]entity.Task, 0),
		mu:    &sync.Mutex{},
	}
}

// Create создание таски в локальном хранилище
func (r *TaskRepository) Create(ctx context.Context, task entity.Task) (id int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks = append(r.tasks, task) // пишем новую таску
	task.Id = r.id                  //выдем таску ID
	r.id++                          // сдвигаем счетчик
	return task.Id, nil
}
