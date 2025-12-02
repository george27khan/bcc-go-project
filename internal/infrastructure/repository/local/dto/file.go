package dto

import "bcc-go-project/internal/domain/entity"

type File struct {
	Id     entity.IdFile
	Data   []byte
	Url    entity.Url
	Error  error
	TaskId entity.IdTask
}

func ToFileDomain(file File) *entity.File {
	return &entity.File{
		Id:    file.Id,
		Data:  file.Data,
		Url:   file.Url,
		Error: file.Error,
	}
}

func ToFileRepo(file entity.File, idTask entity.IdTask) *File {
	return &File{
		Id:     file.Id,
		Data:   file.Data,
		Url:    file.Url,
		Error:  file.Error,
		TaskId: idTask,
	}
}
