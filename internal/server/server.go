package server

import (
	"fmt"
	"github.com/atrian/devmetrics/internal/appconfig"
	"github.com/atrian/devmetrics/internal/server/handlers"
	"github.com/atrian/devmetrics/internal/server/storage"
	"log"
	"net/http"
	"time"
)

type Server struct {
	config  *appconfig.Config
	storage storage.Repository
}

func NewServer() *Server {
	config := appconfig.NewConfig()
	memoryStorage := storage.NewMemoryStorage(config)

	server := Server{
		config:  config,
		storage: memoryStorage,
	}

	return &server
}

func (s *Server) Run() {
	fmt.Printf("Starting server at %v\n", s.config.HTTP.Address)
	defer s.Stop()

	handler := handlers.NewHandler(s.config, s.storage)

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
			log.Fatal(err)
		}
	}

	// запуск сервера, по умолчанию с адресом localhost, порт 8080
	log.Fatal(http.ListenAndServe(s.config.HTTP.Address, handler))
}

func (s *Server) RunMetricsDumpTicker() {
	// запускаем тикер дампа статистики
	dumpMetricsTicker := time.NewTicker(s.config.Server.StoreInterval)

	fmt.Println("Run metrics dump every:", s.config.Server.StoreInterval)

	go func() {
		for dumpTime := range dumpMetricsTicker.C {
			err := s.storage.DumpToFile(s.config.Server.StoreFile)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Metrics dump time:", dumpTime)
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("Dump metrics before shutdown")
	err := s.storage.DumpToFile(s.config.Server.StoreFile)
	if err != nil {
		log.Fatal(err)
	}
}
