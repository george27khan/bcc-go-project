package file

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"sync"
)

//var _ task.FileRepositiry = (*FileRepository)(nil)

type FileRepository struct {
	files []entity.File
	mu    *sync.Mutex
	id    int
}

func NewFileRepository() *FileRepository {
	return &FileRepository{
		files: make([]entity.File, 0),
		mu:    &sync.Mutex{},
	}
}

// Create создание таски в локальном хранилище
func (r *FileRepository) Create(ctx context.Context, file entity.File) (id int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.files = append(r.files, file) // пишем новый файл
	file.Id = r.id                  //выдем файлу ID
	r.id++                          // сдвигаем счетчик
	return file.Id, nil
}
