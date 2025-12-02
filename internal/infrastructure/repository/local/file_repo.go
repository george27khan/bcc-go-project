package local

import (
	"bcc-go-project/internal/domain/entity"
	"bcc-go-project/internal/infrastructure/repository/local/dto"
	"bcc-go-project/internal/usecase"
	"context"
	"sync"
)

var _ usecase.FileRepository = (*FileRepository)(nil)

type FileRepository struct {
	files []dto.File
	mu    *sync.RWMutex
	id    entity.IdFile
}

func NewFileRepository() *FileRepository {
	return &FileRepository{
		files: make([]dto.File, 0),
		mu:    &sync.RWMutex{},
	}
}

// Create создание файла в локальном хранилище
func (r *FileRepository) Create(ctx context.Context, file entity.File, idTask entity.IdTask) (id entity.IdFile, err error) {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	fileRepo := dto.ToFileRepo(file, idTask)
	r.files = append(r.files, *fileRepo) // пишем новый файл
	file.Id = r.id                       //выдем файлу ID
	r.id++                               // сдвигаем счетчик
	return file.Id, nil
}

// UpdateData сохранение скаченного файла
func (r *FileRepository) UpdateData(ctx context.Context, id entity.IdFile, data []byte) error {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.files[id].Data = data
	return nil
}

// UpdateErr сохранение ошибки
func (r *FileRepository) UpdateErr(ctx context.Context, id entity.IdFile, err error) error {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.files[id].Error = err
	return nil
}

// Get получить файл
func (r *FileRepository) Get(ctx context.Context, id entity.IdFile) (*entity.File, error) {
	// завершение операции по контексту
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	fileRepo := r.files[id]
	file := dto.ToFileDomain(fileRepo)
	return file, nil
}

//// UpdateErr сохранение ошибки
//func (r *FileRepository) GetFilesByTaskId(ctx context.Context, id entity.IdTask) ([]*entity.File, error) {
//	// завершение операции по контексту
//	select {
//	case <-ctx.Done():
//		return nil, ctx.Err()
//	default:
//	}
//	files := make([]*entity.File, 0)
//	r.mu.RLock()
//	defer r.mu.RUnlock()
//	for _, file := range r.files {
//		if file.TaskId == id {
//			files = append(files, &file)
//		}
//	}
//	file := r.files[id]
//	return &file, nil
//}
