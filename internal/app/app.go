package app

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
	"go.uber.org/fx"
	"log"
	"net/http"
	"time"
)

func Run() {
	GetApp().Run()
}

func GetApp() *fx.App {
	return fx.New(
		fx.Provide(
			fx.Annotate(
				fileRep.NewTaskRepository,
				fx.As(new(task.TaskFileRepository)),
				fx.As(new(task.GetTaskRepository)),
				fx.As(new(task.CreateTaskRepository)),
			),
			fx.Annotate(task.NewAsyncRunner,
				fx.As(new(task.BackgroundRunner))),
			fx.Annotate(http_loader.NewHttpLoader,
				fx.As(new(task.HttpLoader))),
			fx.Annotate(
				task.NewCreateTaskUseCase,
				fx.As(new(server.TaskCreateUseCase)),
			),
			fx.Annotate(
				task.NewGetTaskUseCase,
				fx.As(new(server.TaskGetUseCase)),
			),
			fx.Annotate(
				task.NewTaskFileUseCase,
				fx.As(new(server.TaskFileUseCase)),
			),
			server.NewTaskServer,
			NewHttpServer),
		fx.Invoke(func(*http.Server) {}),
	)
}

func NewHttpServer(lc fx.Lifecycle, ssi *server.TaskServer, runner task.BackgroundRunner) *http.Server {
	router := chi.NewRouter()
	spec, err := server.GetSwagger()
	if err != nil {
		log.Printf("Ошибка Swagger: %v", err)
		return nil
	}

	// Middleware проверки запросов
	router.Use(mw.OapiRequestValidatorWithOptions(
		spec, // Добавление валидатора сваггера
		&mw.Options{
			Options:      openapi3filter.Options{},
			ErrorHandler: server.SwaggerErrorHandlerFunc, // добавление обработчика ошибок на уровне проверки сваггером
		},
	))
	// Регистрируем все эндпоинты из OpenAPI
	taskSrvStrict := server.NewStrictHandlerWithOptions(
		ssi,
		[]server.StrictMiddlewareFunc{middleware.AddRequestId, middleware.PanicRecover},
		server.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  server.RequestErrorHandlerFunc,
			ResponseErrorHandlerFunc: server.ResponseErrorHandlerFunc,
		},
	)
	server.HandlerFromMux(taskSrvStrict, router)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Printf("Start server on port 8080")
				if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
					log.Printf("listen error: %s", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Start server shutdown ...")
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			//time.Sleep(10 * time.Second)
			if err := srv.Shutdown(ctx); err != nil {
				log.Fatal("Server shutdown err:", err)
				return err
			}
			log.Println("Server exiting")

			runner.(task.AsyncRunner).WgRoot.Wait() // Ожидаем завершение загрузок, теоретически не должно зависнуть т.к. загрузки с таймаутом
			log.Println("Background finish")
			return nil
		},
	})

	return srv
}
