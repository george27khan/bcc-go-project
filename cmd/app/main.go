package main

import (
	"bcc-go-project/internal/app"
)

func main() {
	app.Run()
}

//func main1() {
//	// глобальный контекст для отмены фоновых загрузок при остановке приложения
//	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
//	defer stop()
//
//	// Для ожидания завершения фоновых загрузок при остановке приложения
//	wgBackRun := &sync.WaitGroup{}
//
//	r := chi.NewRouter()
//
//	spec, err := server.GetSwagger()
//	if err != nil {
//		log.Printf("Ошибка Swagger: %v", err)
//		return
//	}
//
//	// Middleware проверки запросов
//	r.Use(mw.OapiRequestValidatorWithOptions(
//		spec, // Добавление валидатора сваггера
//		&mw.Options{
//			Options:      openapi3filter.Options{},
//			ErrorHandler: server.SwaggerErrorHandlerFunc, // добавление обработчика ошибок на уровне проверки сваггером
//		},
//	))
//	taskRep := fileRep.NewTaskRepository()                                  // репозиторий
//	loader := http_loader.NewHttpLoader()                                   // загрузчик
//	runner := &task.AsyncRunner{WgRoot: wgBackRun}                          // запуск загрузки в фоне
//	taskCreateUseCase := task.NewCreateTaskUseCase(taskRep, loader, runner) // юзкейс
//	taskGetUseCase := task.NewGetTaskUseCase(taskRep)                       // юзкейс
//	taskFileUseCase := task.NewTaskFileUseCase(taskRep)                     // юзкейс
//	taskSrv := server.NewTaskServer(taskCreateUseCase, taskGetUseCase, taskFileUseCase)
//
//	// Регистрируем все эндпоинты из OpenAPI
//	taskSrvStrict := server.NewStrictHandlerWithOptions(
//		taskSrv,
//		[]server.StrictMiddlewareFunc{middleware.AddRequestId, middleware.PanicRecover},
//		server.StrictHTTPServerOptions{
//			RequestErrorHandlerFunc:  server.RequestErrorHandlerFunc,
//			ResponseErrorHandlerFunc: server.ResponseErrorHandlerFunc,
//		},
//	)
//	server.HandlerFromMux(taskSrvStrict, r)
//	s := &http.Server{
//		Addr:    ":8080",
//		Handler: r,
//	}
//	go func() {
//		log.Printf("Start server on port 8080")
//		if err := s.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
//			log.Printf("listen error: %s", err)
//		}
//	}()
//
//	//gracefull shutdown
//	<-rootCtx.Done() // ожидание сигнала завершения
//	log.Println("Start gracefull shutdown ...")
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := s.Shutdown(ctx); err != nil {
//		log.Fatal("Server Shutdown err:", err)
//	}
//	log.Println("Server exiting")
//
//	wgBackRun.Wait() // Ожидаем завершение загрузок, теоретически не должно зависнуть т.к. загрузки с таймаутом
//	log.Println("Background stoped")
//
//}
