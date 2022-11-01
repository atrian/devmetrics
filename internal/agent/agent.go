package agent

import (
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/atrian/devmetrics/internal/appconfig/agentconfig"
)

type (
	gauge   float64
	counter int64
)

type Agent struct {
	config  *agentconfig.Config
	metrics *MetricsDics
	logger  *zap.Logger
}

func (a *Agent) Run() {
	a.logger.Info("Agent started",
		zap.Duration("PollInterval", a.config.Agent.PollInterval),
		zap.Duration("ReportInterval", a.config.Agent.ReportInterval),
		zap.String("Server address", a.config.HTTP.Address),
	)

	// запускаем тикер сбора статистики
	refreshStatsTicker := time.NewTicker(a.config.Agent.PollInterval)

	// запускаем тикер отправки статистики
	uploadStatsTicker := time.NewTicker(a.config.Agent.ReportInterval)

	// получаем сигнал из тикеров и запускаем методы сбора и отправки
	for {
		select {
		case refreshTime := <-refreshStatsTicker.C:
			a.logger.Debug("Runtime metrics refresh", zap.Time("time", refreshTime))
			a.RefreshStats()
		case uploadTime := <-uploadStatsTicker.C:
			a.logger.Debug("Metrics upload", zap.Time("time", uploadTime))
			a.UploadStats()
		}
	}
}

func NewAgent() *Agent {
	// подключаем логгер
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Logger init error")
	}
	defer logger.Sync()

	agent := &Agent{
		config:  agentconfig.NewConfig(logger),
		metrics: NewMetricsDicts(),
		logger:  logger,
	}
	return agent
}

func (a *Agent) RefreshStats() {
	a.metrics.updateMetrics()
	a.logger.Info("Runtime stats updated", zap.Int64("PollCount", int64(a.metrics.CounterDict["PollCount"].value)))
}

func (a *Agent) UploadStats() {
	uploader := NewUploader(a.config, a.logger)
	uploader.SendAllStats(a.metrics)
	a.logger.Info("Upload stats")
}
