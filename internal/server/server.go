package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

	server := Server{
		config:  config,
		storage: storage.NewMemoryStorage(config),
	}

	return &server
}

func (s *Server) Run() {
	fmt.Printf("Starting server at %v\n", s.config.HTTP.Address)
	defer s.Stop()

	routes := router.New(handlers.New(s.config, s.storage))

	// STORE_INTERVAL (по умолчанию 300) — интервал времени в секундах,
	// по истечении которого текущие показания сервера сбрасываются на диск
	// (значение 0 — делает запись синхронной).
	if s.config.Server.StoreInterval != 0 {
		s.RunMetricsDumpTicker()
	}

	// RESTORE (по умолчанию true) — булево значение (true/false), определяющее,
	// загружать или нет начальные значения из указанного файла при старте сервера.
	if s.config.Server.Restore {
		err := s.storage.RestoreFromFile(s.config.Server.StoreFile)
		if err != nil {
			fmt.Println(err)
		}
	}

	// запуск сервера, по умолчанию с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(s.config.HTTP.Address, routes))
}

func (s *Server) RunMetricsDumpTicker() {
	// запускаем тикер дампа статистики
	dumpMetricsTicker := time.NewTicker(s.config.Server.StoreInterval)

	fmt.Println("Run metrics dump every:", s.config.Server.StoreInterval)

	go func() {
		for dumpTime := range dumpMetricsTicker.C {
			err := s.storage.DumpToFile(s.config.Server.StoreFile)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Metrics dump time:", dumpTime)
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("Dump metrics before shutdown")
	err := s.storage.DumpToFile(s.config.Server.StoreFile)
	if err != nil {
		fmt.Println(err)
	}
}
