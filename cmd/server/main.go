package main

import (
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

func main() {
	statServer := server.NewServer()
	statServer.Run()
}
