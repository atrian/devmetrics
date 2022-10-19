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
	fmt.Printf("Starting server at %v://%v port:%d\n", s.config.HTTP.Protocol, s.config.HTTP.Server, s.config.HTTP.Port)

	var handler = handlers.NewHandler()

	// запуск сервера с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%v:%d", s.config.HTTP.Server, s.config.HTTP.Port), handler))
}

func NewServer() *Server {
	server := Server{
		config:  appconfig.NewConfig(),
		storage: storage.NewMemoryStorage(),
	}
	return &server
}
