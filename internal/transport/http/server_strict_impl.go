package http

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type DownloadInitUseCase interface {
	CreateTask(ctx context.Context, task entity.Task, files []entity.File) (id int, status string, err error)
}

// StrictServerImpl реализует интерфейс сервера
type StrictServerImpl struct {
	DownloadInit DownloadInitUseCase
}

func NewStrictServerImpl(downloadInit DownloadInitUseCase) *StrictServerImpl {
	return &StrictServerImpl{DownloadInit: downloadInit}
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

func PostDownloadsBadRequest(w http.ResponseWriter, code string, message string, err error) {
	resp := PostDownloads400JSONResponse{
		BadRequestJSONResponse: BadRequestJSONResponse{
			Code:    code,
			Message: fmt.Sprintf("%s: %s", message, err),
		},
	}
	_ = resp.VisitPostDownloadsResponse(w)
	return
}

func PostDownloadsServerError(w http.ResponseWriter, code string, message string, err error) {
	resp := PostDownloads500JSONResponse{
		InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
			Code:    code,
			Message: fmt.Sprintf("%s: %s", message, err),
		},
	}
	_ = resp.VisitPostDownloadsResponse(w)
	return
}

func (s *StrictServerImpl) PostDownloads(ctx context.Context, request PostDownloadsRequestObject) (PostDownloadsResponseObject, error) {
	if err := validate(request.Body); err != nil {
		return &PostDownloads400JSONResponse{
			BadRequestJSONResponse{
				"REQUEST_BODY_ERROR",
				fmt.Sprintf("Ошибка валидации параметров: %s", err),
			},
		}, err
	}
	timeout, err := strToDuration(request.Body.Timeout)
	if err != nil {
		return &PostDownloads400JSONResponse{
			BadRequestJSONResponse{
				"REQUEST_BODY_ERROR",
				fmt.Sprintf("Ошибка обработки таймаута: %s", err),
			},
		}, err
	}
	task := entity.NewTask(timeout)
	files := make([]entity.File, len(request.Body.Files))
	for i, file := range request.Body.Files {
		files[i] = entity.NewFile(file.Url)
	}

	taskId, taskStatus, err := s.DownloadInit.CreateTask(ctx, task, files)
	if err != nil {
		return &PostDownloads500JSONResponse{
			InternalServerErrorJSONResponse{
				"INTERNAL_ERROR",
				fmt.Sprintf("Ошибка  при создании таска на загрузку: %s", err),
			},
		}, err
	}
	return &PostDownloads201JSONResponse{
		Id:     taskId,
		Status: taskStatus,
	}, nil
}

// Реализуем метод конкретного эндпоинта
func (s *StrictServerImpl) GetDownloadsId(ctx context.Context, request GetDownloadsIdRequestObject) (GetDownloadsIdResponseObject, error) {

	return nil, nil
}
