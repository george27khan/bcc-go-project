package main

import (
	"bcc-go-project/internal/infrastructure/adatper/http_loader"
	file_rep "bcc-go-project/internal/infrastructure/repository/local"
	oapi_srv "bcc-go-project/internal/transport/http"
	"bcc-go-project/internal/usecase"
	"github.com/go-chi/chi/v5"
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

	r.Use(middleware.OapiRequestValidator(spec))
	tRep := file_rep.NewTaskRepository()
	loader := http_loader.NewHttpLoader(&http.Client{})
	taskUseCase := usecase.NewTaskUseCase(tRep, loader)
	srv := oapi_srv.NewStrictServerImpl(taskUseCase)

	// Регистрируем все эндпоинты из OpenAPI
	srv1 := oapi_srv.NewStrictHandler(srv, []oapi_srv.StrictMiddlewareFunc{})
	oapi_srv.HandlerFromMux(srv1, r)
	http.ListenAndServe(":8080", r)
}
