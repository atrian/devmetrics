package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/atrian/devmetrics/internal/appconfig"
	"github.com/atrian/devmetrics/internal/server/handlers"
	"github.com/atrian/devmetrics/internal/server/storage"
)

type Server struct {
	config  *appconfig.Config
	storage storage.Repository
}

func (s *Server) Run() {
	fmt.Printf("Starting server at %v\n", s.config.HTTP.Address)

	var handler = handlers.NewHandler()

	// запуск сервера, по умолчанию с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(s.config.HTTP.Address, handler))
}

func NewServer() *Server {
	server := Server{
		config:  appconfig.NewConfig(),
		storage: storage.NewMemoryStorage(),
	}
	return &server
}
