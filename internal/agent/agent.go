package agent

import (
	"fmt"
	"time"

	"github.com/atrian/devmetrics/internal/appconfig/agentconfig"
	"github.com/atrian/devmetrics/pkg/logger"
)

type (
	gauge   float64
	counter int64
)

type Agent struct {
	config  *agentconfig.Config
	metrics *MetricsDics
	logger  logger.Logger
}

func (a *Agent) Run() {

	a.logger.Info(
		fmt.Sprintf("Agent started. PollInterval: %v, ReportInterval: %v, Server address: %v",
			a.config.Agent.PollInterval,
			a.config.Agent.ReportInterval,
			a.config.HTTP.Address))

	// запускаем тикер сбора статистики
	refreshStatsTicker := time.NewTicker(a.config.Agent.PollInterval)

	// запускаем тикер отправки статистики
	uploadStatsTicker := time.NewTicker(a.config.Agent.ReportInterval)

	// получаем сигнал из тикеров и запускаем методы сбора и отправки
	for {
		select {
		case refreshTime := <-refreshStatsTicker.C:
			a.logger.Debug(fmt.Sprintf("Metrics refresh. Time: %v", refreshTime))
			go a.RefreshRuntimeStats()
			go a.RefreshGopsStats()
		case uploadTime := <-uploadStatsTicker.C:
			a.logger.Debug(fmt.Sprintf("Metrics upload. Time: %v", uploadTime))
			go a.UploadStats()
		}
	}
}

func NewAgent() *Agent {
	// подключаем логгер
	agentLogger := logger.NewZapLogger()
	defer agentLogger.Sync()

	agent := &Agent{
		config:  agentconfig.NewConfig(agentLogger),
		metrics: NewMetricsDicts(agentLogger),
		logger:  agentLogger,
	}
	return agent
}

func (a *Agent) RefreshRuntimeStats() {
	a.metrics.updateRuntimeMetrics()

	a.logger.Info(fmt.Sprintf("Runtime stats updated. PollCount: %v",
		int64(a.metrics.CounterDict["PollCount"].value)))
}

func (a *Agent) RefreshGopsStats() {
	a.metrics.updateGopsMetrics()

	a.logger.Info(fmt.Sprintf("Gops stats updated. PollCount: %v",
		int64(a.metrics.CounterDict["PollCount"].value)))
}

func (a *Agent) UploadStats() {
	uploader := NewUploader(a.config, a.logger)
	uploader.SendAllStats(a.metrics)
	a.logger.Info("Upload stats")
}
