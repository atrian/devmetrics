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

type Server struct {
	config  *serverconfig.Config
	storage storage.Repository
	logger  logger.Logger
}

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

func (s *Server) Stop() {
	// выполняем процедуры остановки хранилища
	s.storage.RunOnClose()
}
