package http

import (
	"bcc-go-project/internal/domain/entity"
	rep_err "bcc-go-project/internal/infrastructure/repository/errors"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TaskUseCase interface {
	CreateTask(ctx context.Context, task entity.Task) (id entity.IdTask, status string, err error)
	GetTask(ctx context.Context, id entity.IdTask) (*entity.Task, error)
	GetTaskFile(ctx context.Context, idTask entity.IdTask, idFile entity.IdFile) ([]byte, error)
}

// StrictServerImpl реализует интерфейс сервера
type StrictServerImpl struct {
	TaskUseCase TaskUseCase
}

func NewStrictServerImpl(taskUseCase TaskUseCase) *StrictServerImpl {
	return &StrictServerImpl{TaskUseCase: taskUseCase}
}

func validateURL(u string) error {
	parsed, err := url.ParseRequestURI(u)
	if err != nil {
		return fmt.Errorf("некорректный URL: %s", u)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("URL должен содержать схему и хост: %s", u)
	}
	return nil
}

func validate(req *PostDownloadsJSONRequestBody) error {
	for _, file := range req.Files {
		if err := validateURL(file.Url); err != nil {
			return fmt.Errorf("validate error: %w", err)
		}
	}
	return nil
}

func strToDuration(timeStr string) (time.Duration, error) {
	timeout, err := strconv.Atoi(strings.Trim(timeStr, "s"))
	if err != nil {
		return 0, fmt.Errorf("strToDuration error: %w", err)
	}
	return time.Second * time.Duration(timeout), nil
}

// Реализуем метод конкретного эндпоинта
// POST localhost:8080/downloads
/*
{
"files": [
{"url": "https://google.com"},
{"url": "https://somehost.com/test.pdf"}
],
"timeout": "60s"
}
*/

//func PostDownloadsBadRequest(w http.ResponseWriter, code string, message string, err error) {
//	resp := PostDownloads400JSONResponse{
//		BadRequest400JSONResponse: BadRequest400JSONResponse{
//			Code:    code,
//			Message: fmt.Sprintf("%s: %s", message, err),
//		},
//	}
//	_ = resp.VisitPostDownloadsResponse(w)
//	return
//}
//
//func PostDownloadsServerError(w http.ResponseWriter, code string, message string, err error) {
//	resp := PostDownloads500JSONResponse{
//		InternalServerError500JSONResponse: InternalServerError500JSONResponse{
//			Code:    code,
//			Message: fmt.Sprintf("%s: %s", message, err),
//		},
//	}
//	_ = resp.VisitPostDownloadsResponse(w)
//	return
//}

func (s *StrictServerImpl) PostDownloads(ctx context.Context, request PostDownloadsRequestObject) (PostDownloadsResponseObject, error) {
	if err := validate(request.Body); err != nil {
		return &PostDownloads400JSONResponse{
			BadRequest400JSONResponse{
				BadRequest,
				fmt.Sprintf("PostDownloads: ошибка валидации параметров: %s", err),
			},
		}, err
	}
	timeout, err := strToDuration(request.Body.Timeout)
	if err != nil {
		return &PostDownloads400JSONResponse{
			BadRequest400JSONResponse{
				BadRequest,
				fmt.Sprintf("PostDownloads: ошибка валидации таймаута: %s", err),
			},
		}, err
	}
	urls := make([]entity.Url, len(request.Body.Files))
	for i, url := range request.Body.Files {
		urls[i] = entity.Url(url.Url)
	}
	task := entity.NewTask(timeout, urls)
	taskId, taskStatus, err := s.TaskUseCase.CreateTask(ctx, task)
	if err != nil {
		return &PostDownloads500JSONResponse{
			InternalServerError500JSONResponse{
				InternalServerError,
				fmt.Sprintf("PostDownloads: ошибка  при создании таска на загрузку: %s", err),
			},
		}, err
	}
	return &PostDownloads201JSONResponse{
		Id:     int(taskId),
		Status: taskStatus,
	}, nil
}

// GetDownloadsId
func (s *StrictServerImpl) GetDownloadsId(ctx context.Context, request GetDownloadsIdRequestObject) (GetDownloadsIdResponseObject, error) {
	task, err := s.TaskUseCase.GetTask(ctx, entity.IdTask(request.Id))
	if err != nil {
		if errors.Is(err, rep_err.ErrTaskNotExist) {
			return GetDownloadsId404JSONResponse{NotFound404JSONResponse{
				Code:    NotFound,
				Message: fmt.Errorf("GetDownloadsId: %w", err).Error(),
			}}, nil
		}
	}

	//собираем файлы для ответа
	files := make([]DownloadsIdResponse_Files_Item, len(task.Files))
	for i, file := range task.Files {
		item := DownloadsIdResponse_Files_Item{}
		if file.Error != nil {
			urlErr := UrlErr{
				Url: UrlString(file.Url),
			}
			urlErr.Error.Code = file.Error.Error()
			_ = item.FromUrlErr(urlErr)
		} else {
			urlFile := UrlFile{
				FileId: int(file.Id),
				Url:    UrlString(file.Url),
			}
			_ = item.FromUrlFile(urlFile)
		}
		files[i] = item
	}
	return GetDownloadsId200JSONResponse{
		Files:  files,
		Id:     int(task.Id),
		Status: task.Status,
	}, nil
}

// GetDownloadsIdFilesFileId
// /downloads/0/files/0
func (s *StrictServerImpl) GetDownloadsIdFilesFileId(ctx context.Context, request GetDownloadsIdFilesFileIdRequestObject) (GetDownloadsIdFilesFileIdResponseObject, error) {

	data, err := s.TaskUseCase.GetTaskFile(ctx, entity.IdTask(request.Id), entity.IdFile(request.FileId))
	if err != nil {
		if errors.Is(err, rep_err.ErrTaskNotExist) || errors.Is(err, rep_err.ErrFileNotExist) {
			return GetDownloadsIdFilesFileId404JSONResponse{NotFound404JSONResponse{
				Code:    NotFound,
				Message: fmt.Errorf("GetDownloadsIdFilesFileId: %w", err).Error(),
			}}, nil
		}
	}
	return GetDownloadsIdFilesFileId200ApplicationoctetStreamResponse{
		Body:          bytes.NewReader(data),
		ContentLength: int64(len(data)),
	}, nil
}
