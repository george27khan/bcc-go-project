package main

import (
	"bcc-go-project/internal/infrastructure/adatper/http_loader"
	file_rep "bcc-go-project/internal/infrastructure/repository/local"
	oapi_srv "bcc-go-project/internal/transport/http"
	middlewares "bcc-go-project/internal/transport/http/middlewares"
	"bcc-go-project/internal/usecase/task"
	"github.com/getkin/kin-openapi/openapi3filter"
	chi "github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"log/slog"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	spec, err := oapi_srv.GetSwagger()
	if err != nil {
		slog.Error("Failed to get swagger spec", "error", err)
		return
	}
	// Middleware проверки запросов

	r.Use(middleware.OapiRequestValidatorWithOptions(
		spec, //добавление валидатора свагера
		&middleware.Options{
			Options:      openapi3filter.Options{},
			ErrorHandler: oapi_srv.SwaggerErrorHandlerFunc, // добавление обработчика ошибок на уровне проверки сваггером
		},
	))
	taskRep := file_rep.NewTaskRepository()
	loader := http_loader.NewHttpLoader(&http.Client{})
	taskCreateUseCase := task.NewCreateTaskUseCase(taskRep, loader)
	taskGetUseCase := task.NewGetTaskUseCase(taskRep)
	taskFileUseCase := task.NewTaskFileUseCase(taskRep)
	srv := oapi_srv.NewTaskServer(taskCreateUseCase, taskGetUseCase, taskFileUseCase)

	// Регистрируем все эндпоинты из OpenAPI
	srv1 := oapi_srv.NewStrictHandlerWithOptions(
		srv,
		[]oapi_srv.StrictMiddlewareFunc{middlewares.AddRequestId, middlewares.PanicRecover},
		oapi_srv.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  oapi_srv.RequestErrorHandlerFunc,
			ResponseErrorHandlerFunc: oapi_srv.ResponseErrorHandlerFunc,
		},
	)
	oapi_srv.HandlerFromMux(srv1, r)
	http.ListenAndServe(":8080", r)
}
