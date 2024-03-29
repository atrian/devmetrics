package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/atrian/devmetrics/internal/server"
)

// @Title Metrics storage API
// @Description Сервис хранения метрик и счетчиков.
// @Version 1.0

// @Host localhost:8080
// @BasePath /

// @Tag.name Info
// @Tag.description "Группа запросов состояния сервиса"

// @Tag.name Metrics
// @Tag.description "Группа для работы с данными метрик"

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	statServer := server.NewServer()
	statServer.Run(ctx)
}
