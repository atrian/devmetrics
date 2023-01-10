// Package server - серверная часть приложения по сбору метрик.
// Принимает метрики в JSON формате, сохраняет в In Memory или PostrgeSQL хранилище
// Реализована проверка подписи метрик.
// Подключено профилирование, см. конфигурацию serverconfig.ServerConfig
package server

import (
	"fmt"
	"log"
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
	storage storage.IRepository
	logger  logger.ILogger
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
	var appStorage storage.IRepository

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
func (s *Server) Run() {

	s.logger.Info(fmt.Sprintf("Starting server @ %v", s.config.HTTP.Address))
	defer s.Stop()

	routes := router.New(handlers.New(s.config, s.storage, s.logger))
	// Разрешаем роуты профайлера, если разрешено конфигурацией
	if s.config.Server.ProfileApp {
		routes.Mount("/debug", middleware.Profiler())
	}

	// выполняем стартовые процедуры для хранилища
	s.storage.RunOnStart()

	// запуск сервера, по умолчанию с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(s.config.HTTP.Address, routes))
}

// Stop остановка сервера, завершение работы с хранилищем
func (s *Server) Stop() {
	// выполняем процедуры остановки хранилища
	s.storage.RunOnClose()
}
