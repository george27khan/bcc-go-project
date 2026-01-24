package server

import (
	"bcc-go-project/internal/domain/entity"
	repErr "bcc-go-project/internal/infrastructure/repository/errors_repo"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//go:generate mockgen -package server -source=task_server.go -destination=mock_task_server.go
var (
	statusMapping = map[entity.Status]TaskStatus{
		entity.TaskStatusProcess: PROCESS,
		entity.TaskStatusDone:    DONE,
	}
	errFileMapping = map[entity.Error]UrlErrErrorCode{
		entity.FileErr:        UrlErrErrorCodeERROR,
		entity.FileErrTimeout: UrlErrErrorCodeTIMEOUT,
	}
)
var _ StrictServerInterface = (*TaskServer)(nil)

type TaskCreateUseCase interface {
	CreateTask(ctx context.Context, task entity.Task) (id entity.IdTask, status entity.Status, err error)
}
type TaskGetUseCase interface {
	GetTask(ctx context.Context, id entity.IdTask) (*entity.Task, error)
}
type TaskFileUseCase interface {
	GetTaskFile(ctx context.Context, idTask entity.IdTask, idFile entity.IdFile) ([]byte, error)
}

// TaskServer реализует интерфейс сервера
type TaskServer struct {
	TaskCreateUseCase TaskCreateUseCase
	TaskGetUseCase    TaskGetUseCase
	TaskFileUseCase   TaskFileUseCase
}

func NewTaskServer(taskCreateUseCase TaskCreateUseCase, taskGetUseCase TaskGetUseCase, taskFileUseCase TaskFileUseCase) *TaskServer {
	return &TaskServer{
		TaskCreateUseCase: taskCreateUseCase,
		TaskGetUseCase:    taskGetUseCase,
		TaskFileUseCase:   taskFileUseCase,
	}
}

func resp400(msg string, err error) BadRequest400JSONResponse {
	return BadRequest400JSONResponse{
		Code:    ErrorCodeBADREQUEST,
		Message: fmt.Errorf("%s: %w", msg, err).Error(),
	}
}

func resp404(msg string, err error) NotFound404JSONResponse {
	return NotFound404JSONResponse{
		Code:    ErrorCodeNOTFOUND,
		Message: fmt.Errorf("%s: %w", msg, err).Error(),
	}
}

func resp500(msg string, err error) InternalServerError500JSONResponse {
	return InternalServerError500JSONResponse{
		Code:    ErrorCodeINTERNALSERVERERROR,
		Message: fmt.Errorf("%s: %w", msg, err).Error(),
	}
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

// PostDownloads запуск процесса загрузки
func (s *TaskServer) PostDownloads(ctx context.Context, request PostDownloadsRequestObject) (PostDownloadsResponseObject, error) {
	if err := validate(request.Body); err != nil {
		return PostDownloads400JSONResponse{
			resp400("PostDownloads: ошибка валидации параметров", err)}, nil
	}
	timeout, err := strToDuration(request.Body.Timeout)
	//if err != nil {
	//	return PostDownloads400JSONResponse{
	//			resp400("PostDownloads: ошибка валидации таймаута", err)},
	//		nil
	//}
	urls := make([]entity.Url, len(request.Body.Files))
	for i, urlTask := range request.Body.Files {
		urls[i] = entity.Url(urlTask.Url)
	}
	task := entity.NewTask(timeout, urls)
	taskId, taskStatus, err := s.TaskCreateUseCase.CreateTask(ctx, task)
	if err != nil {
		return PostDownloads500JSONResponse{
				resp500("PostDownloads", err)},
			nil
	}
	return PostDownloads201JSONResponse{
		Id:     int(taskId),
		Status: statusMapping[taskStatus],
	}, nil
}

// GetDownloadsId получение таска
func (s *TaskServer) GetDownloadsId(ctx context.Context, request GetDownloadsIdRequestObject) (GetDownloadsIdResponseObject, error) {
	task, err := s.TaskGetUseCase.GetTask(ctx, entity.IdTask(request.Id))
	if err != nil {
		if errors.Is(err, repErr.ErrTaskNotExist) {
			return GetDownloadsId404JSONResponse{resp404("GetDownloadsId", err)}, nil
		}
	}
	//собираем файлы для ответа
	files := make([]DownloadsIdResponse_Files_Item, len(task.Files))
	for i, file := range task.Files {
		item := DownloadsIdResponse_Files_Item{}
		if file.Error != "" {
			urlErr := UrlErr{
				Url: UrlString(file.Url),
			}
			urlErr.Error.Code = errFileMapping[file.Error]
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
		Status: statusMapping[task.Status],
	}, nil
}

// GetDownloadsIdFilesFileId
// /downloads/0/files/0
func (s *TaskServer) GetDownloadsIdFilesFileId(ctx context.Context, request GetDownloadsIdFilesFileIdRequestObject) (GetDownloadsIdFilesFileIdResponseObject, error) {
	data, err := s.TaskFileUseCase.GetTaskFile(ctx, entity.IdTask(request.Id), entity.IdFile(request.FileId))
	if err != nil {
		if errors.Is(err, repErr.ErrTaskNotExist) || errors.Is(err, repErr.ErrFileNotExist) {
			return GetDownloadsIdFilesFileId404JSONResponse{resp404("GetDownloadsIdFilesFileId", err)}, nil
		}
	}
	return GetDownloadsIdFilesFileId200ApplicationoctetStreamResponse{
		Body:          bytes.NewReader(data),
		ContentLength: int64(len(data)),
	}, nil
}
