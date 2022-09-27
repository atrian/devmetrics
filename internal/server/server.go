package server

import (
	"fmt"
	"github.com/atrian/devmetrics/internal/appconfig"
	"github.com/atrian/devmetrics/internal/server/storage"
)

type Server struct {
	config  *appconfig.Config
	storage storage.Repository
}

func (s *Server) Run() {
	fmt.Printf("Starting server at %v port:%d\n", s.config.HTTP.Server, s.config.HTTP.Port)
}

func NewServer() *Server {
	server := Server{
		config:  appconfig.NewConfig(),
		storage: storage.NewMemoryStorage(),
	}

	return &server
}
