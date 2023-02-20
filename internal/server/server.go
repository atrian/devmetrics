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
	"github.com/atrian/devmetrics/internal/server/grpcHandlers"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/server/handlers"
	"github.com/atrian/devmetrics/internal/server/router"
	"github.com/atrian/devmetrics/internal/server/storage"
	"github.com/atrian/devmetrics/pkg/logger"
	pb "github.com/atrian/devmetrics/proto"
)

// Server основное приложение сервера.
// Доступен профайлер при ProfileApp == true в serverconfig.ServerConfig
type Server struct {
	config  *serverconfig.Config
	storage storage.Repository
	logger  logger.Logger
	web     http.Server
	grpc    *grpc.Server
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
	go s.graceShutdownWatcher(ctx, graceShutdown)

	// выполняем стартовые процедуры для хранилища
	s.storage.RunOnStart()

	go s.runWebServer()
	go s.runGRPCServer(ctx)

	<-graceShutdown
}

func (s *Server) runGRPCServer(ctx context.Context) {
	s.logger.Info(fmt.Sprintf("Starting GRPC server @ %v", s.config.Transport.AddressGRPC))

	// определяем порт для сервера
	listen, err := net.Listen("tcp", s.config.Transport.AddressGRPC)
	if err != nil {
		s.logger.Fatal("GRPC net.Listen error", err)
	}
	// создаём gRPC-сервер без зарегистрированной службы
	s.grpc = grpc.NewServer()
	// регистрируем сервис
	ms := grpcHandlers.NewMetricServer(s.storage, s.logger)
	pb.RegisterDevMetricsServer(s.grpc, ms)

	s.logger.Info("GRPC server started")

	// получаем запрос gRPC
	if err = s.grpc.Serve(listen); err != nil {
		log.Fatal(err)
	}
}

// runWebServer конфигурирование и запуск веб сервера
func (s *Server) runWebServer() {
	// слайс для кастомных middlewares
	var customMiddlewares []func(next http.Handler) http.Handler

	s.logger.Info(fmt.Sprintf("Starting server @ %v", s.config.Transport.AddressHTTP))

	api := handlers.New(s.config, s.storage, s.logger)
	routes := router.New(api, customMiddlewares, s.config)
	// Разрешаем роуты профайлера, если разрешено конфигурацией
	if s.config.Server.ProfileApp {
		routes.Mount("/debug", middleware.Profiler())
	}

	// запуск сервера, по умолчанию с адресом localhost, порт 8080
	s.web.Addr = s.config.Transport.AddressHTTP
	s.web.Handler = routes
	err := s.web.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		s.logger.Fatal("Server ListenAndServe err", err)
	}
}

func (s *Server) graceShutdownWatcher(ctx context.Context, grace chan struct{}) {
	<-ctx.Done()
	go func() {
		s.Stop(grace)
	}()
}

// Stop остановка сервера, завершение работы с хранилищем
func (s *Server) Stop(grace chan struct{}) {
	defer close(grace)

	// выполняем остановку WEB сервера
	err := s.web.Shutdown(context.Background())
	if err != nil {
		s.logger.Error("Server stop error", err)
	}
	s.logger.Info("WEB server shutdown gracefully")

	// GracefulStop stops the gRPC server gracefully. It stops the server from
	// accepting new connections and RPCs and blocks until all the pending RPCs are
	// finished.
	s.grpc.GracefulStop()
	s.logger.Info("GRPC server shutdown gracefully")

	// выполняем процедуры остановки хранилища
	s.storage.RunOnClose()
	s.logger.Info("Storage shutdown gracefully")
}
