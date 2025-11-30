package main

import (
	"bcc-go-project/internal/api"
	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"log/slog"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	spec, err := api.GetSwagger()
	if err != nil {
		slog.Error("Failed to get swagger spec", "error", err)
		return
	}
	// Middleware проверки запросов

	r.Use(middleware.OapiRequestValidator(spec))

	server := &api.ServerImpl{}

	// Регистрируем все эндпоинты из OpenAPI
	api.HandlerFromMux(server, r)

	http.ListenAndServe(":8080", r)
}
