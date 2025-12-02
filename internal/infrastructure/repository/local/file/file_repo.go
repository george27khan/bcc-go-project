package file

import (
	"bcc-go-project/internal/domain/entity"
	"bcc-go-project/internal/usecase/task_usecase"
	"context"
	"sync"
)

var _ task_usecase.FileRepository = (*FileRepository)(nil)

type FileRepository struct {
	files []entity.File
	mu    *sync.RWMutex
	id    int
}

func NewFileRepository() *FileRepository {
	return &FileRepository{
		files: make([]entity.File, 0),
		mu:    &sync.RWMutex{},
	}
}

// Create создание файла в локальном хранилище
func (r *FileRepository) Create(ctx context.Context, file entity.File) (id int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.files = append(r.files, file) // пишем новый файл
	file.Id = r.id                  //выдем файлу ID
	r.id++                          // сдвигаем счетчик
	return file.Id, nil
}

func (r *FileRepository) UpdateData(ctx context.Context, id int, data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.files[id].Data = data
	return nil
}

func (r *FileRepository) UpdateErr(ctx context.Context, id int, err error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.files[id].Error = err
	return nil
}
