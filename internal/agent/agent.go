package agent

import (
	"fmt"
	"time"

	"github.com/atrian/devmetrics/internal/appconfig/agentconfig"
)

type (
	gauge   float64
	counter int64
)

type Agent struct {
	config  *agentconfig.Config
	metrics *MetricsDics
}

func (a *Agent) Run() {
	fmt.Printf("Agent started with PollInterval: %v, ReportInterval: %v, Server address: %v\n",
		a.config.Agent.PollInterval, a.config.Agent.ReportInterval, a.config.HTTP.Address)

	// запускаем тикер сбора статистики
	refreshStatsTicker := time.NewTicker(a.config.Agent.PollInterval)

	// запускаем тикер отправки статистики
	uploadStatsTicker := time.NewTicker(a.config.Agent.ReportInterval)

	// получаем сигнал из тикеров и запускаем методы сбора и отправки
	for {
		select {
		case refreshTime := <-refreshStatsTicker.C:
			fmt.Println("refresh time:", refreshTime)
			a.RefreshStats()
		case uploadTime := <-uploadStatsTicker.C:
			fmt.Println("upload time", uploadTime)
			a.UploadStats()
		}
	}
}

func NewAgent() *Agent {
	agent := &Agent{
		config:  agentconfig.NewConfig(),
		metrics: NewMetricsDicts(),
	}
	return agent
}

func (a *Agent) RefreshStats() {
	a.metrics.updateMetrics()
	fmt.Println("Runtime stats updated")
	fmt.Println(a.metrics.CounterDict["PollCount"].value)
}

func (a *Agent) UploadStats() {
	uploader := NewUploader(a.config)
	uploader.SendAllStats(a.metrics)
	fmt.Println("Upload stats")
}
