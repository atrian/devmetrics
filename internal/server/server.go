package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/server/handlers"
	"github.com/atrian/devmetrics/internal/server/router"
	"github.com/atrian/devmetrics/internal/server/storage"
)

type Server struct {
	config  *serverconfig.Config
	storage storage.Repository
}

func NewServer() *Server {
	config := serverconfig.NewServerConfig()

	var appStorage storage.Repository
	if config.Server.DBDSN == "" {
		appStorage = storage.NewMemoryStorage(config)
	} else {
		appStorage = storage.NewPgSQLStorage(config)
	}

	server := Server{
		config:  config,
		storage: appStorage,
	}

	return &server
}

func (s *Server) Run() {
	fmt.Printf("Starting server at %v\n", s.config.HTTP.Address)
	defer s.Stop()

	routes := router.New(handlers.New(s.config, s.storage))

	// выполняем стартовые процедуры для хранилища
	s.storage.RunOnStart()

	// запуск сервера, по умолчанию с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(s.config.HTTP.Address, routes))
}

func (s *Server) Stop() {
	// выполняем процедуры остановки хранилища
	s.storage.RunOnClose()
}
