// Package server - серверная часть приложения по сбору метрик.
// Принимает метрики в JSON формате, сохраняет в In Memory или PostrgeSQL хранилище
// Реализована проверка подписи метрик.
// Реализовано асинхронное шифрование метрик
// Подключено профилирование, см. конфигурацию serverconfig.ServerConfig
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/server/handlers"
	"github.com/atrian/devmetrics/internal/server/router"
	"github.com/atrian/devmetrics/internal/server/storage"
	"github.com/atrian/devmetrics/pkg/logger"
)

// Server основное приложение сервера.
// Доступен профайлер при ProfileApp == true в serverconfig.ServerConfig
type Server struct {
	config  *serverconfig.Config
	storage storage.Repository
	logger  logger.Logger
	web     http.Server
}

// NewServer возвращает указатель на Server сконфигурированный со всеми зависимостями:
// logger - используется ZAP
// storage - в зависимости от конфигурации, In Memory или PgSQL
// config - собранынй серверный конфиг serverconfig.Config с учетом флагов и переменных окружения
func NewServer() *Server {
	// подключаем логгер
	serverLogger := logger.NewZapLogger()
	defer serverLogger.Sync()

	// подключаем конфиги
	config := serverconfig.NewServerConfig(serverLogger)

	// подключаем storage
	var appStorage storage.Repository

	if config.Server.DBDSN == "" {
		serverLogger.Info("Loading memory storage")
		appStorage = storage.NewMemoryStorage(config, serverLogger)
	} else {
		serverLogger.Info("Loading PGSQL storage")
		var err error
		appStorage, err = storage.NewPgSQLStorage(config, serverLogger)
		if err != nil {
			serverLogger.Error("Loading PGSQL storage", err)
		}
	}

	server := Server{
		config:  config,
		storage: appStorage,
		logger:  serverLogger,
	}

	return &server
}

// Run запуск сервера по указанному адресу.
// если параметр serverconfig.ServerConfig - ProfileApp == true
// доступны хендлеры PPROF
//
//	r.HandleFunc("/pprof/*", pprof.Index)
//	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
//	r.HandleFunc("/pprof/profile", pprof.Profile)
//	r.HandleFunc("/pprof/symbol", pprof.Symbol)
//	r.HandleFunc("/pprof/trace", pprof.Trace)
//	r.HandleFunc("/vars", expVars)
//
//	r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
//	r.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
//	r.Handle("/pprof/mutex", pprof.Handler("mutex"))
//	r.Handle("/pprof/heap", pprof.Handler("heap"))
//	r.Handle("/pprof/block", pprof.Handler("block"))
//	r.Handle("/pprof/allocs", pprof.Handler("allocs"))
func (s *Server) Run(ctx context.Context) {
	graceShutdown := make(chan struct{})
	go s.graceWatcher(ctx, graceShutdown)

	// слайс для кастомных middlewares
	var customMiddlewares []func(next http.Handler) http.Handler

	s.logger.Info(fmt.Sprintf("Starting server @ %v", s.config.HTTP.Address))

	api := handlers.New(s.config, s.storage, s.logger)
	routes := router.New(api, customMiddlewares, s.config)
	// Разрешаем роуты профайлера, если разрешено конфигурацией
	if s.config.Server.ProfileApp {
		routes.Mount("/debug", middleware.Profiler())
	}

	// выполняем стартовые процедуры для хранилища
	s.storage.RunOnStart()

	// запуск сервера, по умолчанию с адресом localhost, порт 8080
	s.web.Addr = s.config.HTTP.Address
	s.web.Handler = routes
	err := s.web.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		s.logger.Fatal("Server ListenAndServe err", err)
	}
	<-graceShutdown
}

func (s *Server) graceWatcher(ctx context.Context, grace chan struct{}) {
	<-ctx.Done()
	go func() {
		s.Stop(grace)
	}()
}

// Stop остановка сервера, завершение работы с хранилищем
func (s *Server) Stop(grace chan struct{}) {
	defer close(grace)

	// выполняем остановку сервера
	err := s.web.Shutdown(context.Background())
	if err != nil {
		s.logger.Error("Server stop error", err)
	}
	s.logger.Info("Server shutdown gracefully")

	// выполняем процедуры остановки хранилища
	s.storage.RunOnClose()
	s.logger.Info("Storage shutdown gracefully")
}
