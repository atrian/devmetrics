package agent

import (
	"fmt"
	"runtime"
	"time"
)

type (
	gauge   float64
	counter int64
)

type Agent struct {
	config  *Config
	metrics *MetricsDics
}

func (a *Agent) Run() {
	fmt.Println("Agent started")

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
		config:  NewConfig(),
		metrics: NewMetricsDicts(),
	}
	return agent
}

func (a *Agent) RefreshStats() {
	// получаем данные мониторинга
	runtimeStats := runtime.MemStats{}

	// обновляем данные мониторинга по списку
	runtime.ReadMemStats(&runtimeStats)

	for _, metric := range a.metrics.GaugeDict {
		metric.value = metric.initValue(&runtimeStats)
	}
	for _, ct := range a.metrics.CounterDict {
		ct.value = ct.initValue(ct)
	}

	fmt.Println("Runtime stats updated")
}

func (a *Agent) UploadStats() {
	fmt.Println("Upload stats", a.buildStatUploadUrl("Type", "Title", "value"))
}

func (a *Agent) buildStatUploadUrl(metricType string, metricTitle string, metricValue string) string {
	return fmt.Sprintf(a.config.Http.UrlTemplate, a.config.Http.Server, a.config.Http.Port, metricType, metricTitle, metricValue)
}
