package main

import (
	"bcc-go-project/internal/infrastructure/adatper/http_loader"
	file_rep "bcc-go-project/internal/infrastructure/repository/local"
	"bcc-go-project/internal/transport/http/server"
	"bcc-go-project/internal/transport/http/server/middleware"
	"bcc-go-project/internal/usecase/task"
	"context"
	"github.com/getkin/kin-openapi/openapi3filter"
	chi "github.com/go-chi/chi/v5"
	mw "github.com/oapi-codegen/nethttp-middleware"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	wgBackRun := &sync.WaitGroup{}
	defer stop()
	r := chi.NewRouter()

	spec, err := server.GetSwagger()
	if err != nil {
		slog.Error("Failed to get swagger spec", "error", err)
		return
	}
	// Middleware проверки запросов

	r.Use(mw.OapiRequestValidatorWithOptions(
		spec, //добавление валидатора свагера
		&mw.Options{
			Options:      openapi3filter.Options{},
			ErrorHandler: server.SwaggerErrorHandlerFunc, // добавление обработчика ошибок на уровне проверки сваггером
		},
	))
	taskRep := file_rep.NewTaskRepository()
	loader := http_loader.NewHttpLoader(&http.Client{})
	runner := &task.AsyncRunner{WgRoot: wgBackRun}
	taskCreateUseCase := task.NewCreateTaskUseCase(taskRep, loader, runner, rootCtx)
	taskGetUseCase := task.NewGetTaskUseCase(taskRep)
	taskFileUseCase := task.NewTaskFileUseCase(taskRep)
	taskSrv := server.NewTaskServer(taskCreateUseCase, taskGetUseCase, taskFileUseCase)

	// Регистрируем все эндпоинты из OpenAPI
	taskSrvStrict := server.NewStrictHandlerWithOptions(
		taskSrv,
		[]server.StrictMiddlewareFunc{middleware.AddRequestId, middleware.PanicRecover},
		server.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  server.RequestErrorHandlerFunc,
			ResponseErrorHandlerFunc: server.ResponseErrorHandlerFunc,
		},
	)
	server.HandlerFromMux(taskSrvStrict, r)
	s := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		log.Printf("Start server on port 8080")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	//gracefull shutdown
	<-rootCtx.Done()
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown err:", err)
	}
	log.Println("Server exiting")
	wgBackRun.Wait()
	log.Println("Background stoped")

}
