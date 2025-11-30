package api

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type DownloadInitUseCase interface {
	CreateTask(ctx context.Context, task entity.Task, files []entity.File) (id int, err error)
}

// ServerImpl реализует интерфейс сервера
type ServerImpl struct {
	StrictServerInterface // встраиваем базовую реализацию
	DownloadInit          DownloadInitUseCase
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

func validate(req DownloadsRequest) error {
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

func badRequest(w http.ResponseWriter, code string, err error) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: err.Error(),
	})
}

func (s *ServerImpl) PostDownloads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqBody DownloadsRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Code:    "",
			Message: err.Error(),
		})
		return
	}
	if err := validate(reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Code:    "REQUEST_BODY_ERROR",
			Message: err.Error(),
		})
		return
	}
	timeout, err := strToDuration(reqBody.Timeout)
	if err != nil {
		badRequest(w, "REQUEST_BODY_ERROR", err)
		return
	}
	task := entity.NewTask(timeout)
	files := make([]entity.File, len(reqBody.Files))
	for _, file := range reqBody.Files {
		files = append(files, entity.NewFile(file.Url))
	}

	_, _ = s.DownloadInit.CreateTask(ctx, task, files)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Hello World")
}

// Реализуем метод конкретного эндпоинта
func (s *ServerImpl) GetDownloadsId(w http.ResponseWriter, r *http.Request, id int) {

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("")
}
