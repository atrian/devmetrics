package server

import (
	"fmt"
	"github.com/atrian/devmetrics/internal/appconfig"
)

type Server struct {
	config *appconfig.Config
}

func (s *Server) Run() {
	fmt.Printf("Starting server at %v port:%d\n", s.config.HTTP.Server, s.config.HTTP.Port)
}

func NewServer() *Server {
	server := Server{
		config: appconfig.NewConfig(),
	}

	return &server
}
