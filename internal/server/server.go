package server

import (
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/server/handlers"
	"github.com/atrian/devmetrics/internal/server/router"
	"github.com/atrian/devmetrics/internal/server/storage"
)

type Server struct {
	config  *serverconfig.Config
	storage storage.Repository
	logger  *zap.Logger
}

func NewServer() *Server {
	// подключаем логгер
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Logger init error")
	}
	defer logger.Sync()

	// подключаем конфиги
	config := serverconfig.NewServerConfig(logger)

	// подключаем storage
	var appStorage storage.Repository
	if config.Server.DBDSN == "" {
		logger.Info("Loading memory storage")
		appStorage = storage.NewMemoryStorage(config, logger)
	} else {
		logger.Info("Loading PGSQL storage")
		appStorage, err = storage.NewPgSQLStorage(config, logger)
	}

	server := Server{
		config:  config,
		storage: appStorage,
		logger:  logger,
	}

	return &server
}

func (s *Server) Run() {
	s.logger.Info("Starting server", zap.String("address", s.config.HTTP.Address))
	defer s.Stop()

	routes := router.New(handlers.New(s.config, s.storage, s.logger))

	// выполняем стартовые процедуры для хранилища
	s.storage.RunOnStart()

	// запуск сервера, по умолчанию с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(s.config.HTTP.Address, routes))
}

func (s *Server) Stop() {
	// выполняем процедуры остановки хранилища
	s.storage.RunOnClose()
}
