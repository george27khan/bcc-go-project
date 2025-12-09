package main

import (
	"bcc-go-project/internal/infrastructure/adatper/http_loader"
	file_rep "bcc-go-project/internal/infrastructure/repository/local"
	"bcc-go-project/internal/transport/http/server"
	"bcc-go-project/internal/transport/http/server/middlewares"
	"bcc-go-project/internal/usecase/task"
	"github.com/getkin/kin-openapi/openapi3filter"
	chi "github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"log/slog"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	spec, err := server.GetSwagger()
	if err != nil {
		slog.Error("Failed to get swagger spec", "error", err)
		return
	}
	// Middleware проверки запросов

	r.Use(middleware.OapiRequestValidatorWithOptions(
		spec, //добавление валидатора свагера
		&middleware.Options{
			Options:      openapi3filter.Options{},
			ErrorHandler: server.SwaggerErrorHandlerFunc, // добавление обработчика ошибок на уровне проверки сваггером
		},
	))
	taskRep := file_rep.NewTaskRepository()
	loader := http_loader.NewHttpLoader(&http.Client{})
	taskCreateUseCase := task.NewCreateTaskUseCase(taskRep, loader)
	taskGetUseCase := task.NewGetTaskUseCase(taskRep)
	taskFileUseCase := task.NewTaskFileUseCase(taskRep)
	srv := server.NewTaskServer(taskCreateUseCase, taskGetUseCase, taskFileUseCase)

	// Регистрируем все эндпоинты из OpenAPI
	srv1 := server.NewStrictHandlerWithOptions(
		srv,
		[]server.StrictMiddlewareFunc{middlewares.AddRequestId, middlewares.PanicRecover},
		server.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  server.RequestErrorHandlerFunc,
			ResponseErrorHandlerFunc: server.ResponseErrorHandlerFunc,
		},
	)
	server.HandlerFromMux(srv1, r)
	http.ListenAndServe(":8080", r)
}
