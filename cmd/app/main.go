package main

import (
	"bcc-go-project/internal/infrastructure/adatper/http_loader"
	fileRep "bcc-go-project/internal/infrastructure/repository/local"
	"bcc-go-project/internal/transport/http/server"
	"bcc-go-project/internal/transport/http/server/middleware"
	"bcc-go-project/internal/usecase/task"
	"context"
	"errors"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	mw "github.com/oapi-codegen/nethttp-middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// глобальный контекст для отмены фоновых загрузок при остановке приложения
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	// Для ожидания завершения фоновых загрузок при остановке приложения
	wgBackRun := &sync.WaitGroup{}

	r := chi.NewRouter()

	spec, err := server.GetSwagger()
	if err != nil {
		log.Printf("Ошибка Swagger: %v", err)
		return
	}

	// Middleware проверки запросов
	r.Use(mw.OapiRequestValidatorWithOptions(
		spec, // Добавление валидатора сваггера
		&mw.Options{
			Options:      openapi3filter.Options{},
			ErrorHandler: server.SwaggerErrorHandlerFunc, // добавление обработчика ошибок на уровне проверки сваггером
		},
	))
	taskRep := fileRep.NewTaskRepository()                                           // репозиторий
	loader := http_loader.NewHttpLoader(&http.Client{})                              // загрузчик
	runner := &task.AsyncRunner{WgRoot: wgBackRun}                                   // запуск загрузки в фоне
	taskCreateUseCase := task.NewCreateTaskUseCase(taskRep, loader, runner, rootCtx) // юзкейс
	taskGetUseCase := task.NewGetTaskUseCase(taskRep)                                // юзкейс
	taskFileUseCase := task.NewTaskFileUseCase(taskRep)                              // юзкейс
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
		if err := s.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen error: %s", err)
		}
	}()

	//gracefull shutdown
	<-rootCtx.Done() // ожидание сигнала завершения
	log.Println("Start gracefull shutdown ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown err:", err)
	}
	log.Println("Server exiting")

	wgBackRun.Wait() // Ожидаем завершение загрузок, теоретически не должно зависнуть т.к. загрузки с таймаутом
	log.Println("Background stoped")

}
