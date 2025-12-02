package http

//
//import (
//	"bcc-go-project/internal/domain/entity"
//	"context"
//	"encoding/json"
//	"fmt"
//	"net/http"
//	"net/url"
//	"strconv"
//	"strings"
//	"time"
//)
//
//type DownloadInitUseCase interface {
//	CreateTask(ctx context.Context, task_usecase entity.Task, files []entity.File) (id int, status string, err error)
//}
//
//// ServerImpl реализует интерфейс сервера
//type ServerImpl struct {
//	//StrictServerInterface // встраиваем базовую реализацию
//	DownloadInit DownloadInitUseCase
//}
//
//func NewServerImpl(downloadInit DownloadInitUseCase) *ServerImpl {
//	return &ServerImpl{DownloadInit: downloadInit}
//}
//
//func validateURL(u string) error {
//	parsed, err := url.ParseRequestURI(u)
//	if err != nil {
//		return fmt.Errorf("некорректный URL: %s", u)
//	}
//	if parsed.Scheme == "" || parsed.Host == "" {
//		return fmt.Errorf("URL должен содержать схему и хост: %s", u)
//	}
//	return nil
//}
//
//func validate(req DownloadsRequest) error {
//	for _, file := range req.Files {
//		if err := validateURL(file.Url); err != nil {
//			return fmt.Errorf("validate error: %w", err)
//		}
//	}
//	return nil
//}
//
//func strToDuration(timeStr string) (time.Duration, error) {
//	timeout, err := strconv.Atoi(strings.Trim(timeStr, "s"))
//	if err != nil {
//		return 0, fmt.Errorf("strToDuration error: %w", err)
//	}
//	return time.Second * time.Duration(timeout), nil
//}
//
//// Реализуем метод конкретного эндпоинта
//// POST localhost:8080/downloads
///*
//{
//"files": [
//{"url": "https://google.com"},
//{"url": "https://somehost.com/test.pdf"}
//],
//"timeout": "60s"
//}
//*/
//
//func PostDownloadsBadRequest(w http.ResponseWriter, code string, message string, err error) {
//	resp := PostDownloads400JSONResponse{
//		BadRequestJSONResponse: BadRequestJSONResponse{
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
//		InternalServerErrorJSONResponse: InternalServerErrorJSONResponse{
//			Code:    code,
//			Message: fmt.Sprintf("%s: %s", message, err),
//		},
//	}
//	_ = resp.VisitPostDownloadsResponse(w)
//	return
//}
//
//func (s *ServerImpl) PostDownloads(w http.ResponseWriter, r *http.Request) {
//	ctx := r.Context()
//	var reqBody DownloadsRequest
//	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
//		PostDownloadsBadRequest(w, "REQUEST_BODY_ERROR", "Ошибка декодирования тела запроса", err)
//		return
//	}
//	if err := validate(reqBody); err != nil {
//		PostDownloadsBadRequest(w, "REQUEST_BODY_ERROR", "Ошибка валидации параметров", err)
//		return
//	}
//	timeout, err := strToDuration(reqBody.Timeout)
//	if err != nil {
//		PostDownloadsBadRequest(w, "REQUEST_BODY_ERROR", "Ошибка обработки таймаута", err)
//		return
//	}
//	task_usecase := entity.NewTask(timeout)
//	files := make([]entity.File, len(reqBody.Files))
//	for i, file := range reqBody.Files {
//		files[i] = entity.NewFile(file.Url)
//	}
//
//	taskId, taskStatus, err := s.DownloadInit.CreateTask(ctx, task_usecase, files)
//	if err != nil {
//		PostDownloadsServerError(w, "INTERNAL_ERROR", "Ошибка  при создании таска на загрузку", err)
//		return
//	}
//
//	resp := PostDownloads201JSONResponse{
//		Id:     taskId,
//		Status: taskStatus,
//	}
//	_ = resp.VisitPostDownloadsResponse(w)
//}
//
//// Реализуем метод конкретного эндпоинта
//func (s *ServerImpl) GetDownloadsId(w http.ResponseWriter, r *http.Request, id int) {
//
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode("")
//}
