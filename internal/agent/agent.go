// Package agent - клиентская часть приложения по сбору метрик.
// собирает фиксированный список метрик и отправляет на сервер
// Интервалы сбора метрик и отправки настраиваются.
// Данные отправляются в формате JSON в пакетном режиме, применяется Gzip сжатие.
// В приложении доступен профилировщик
package agent

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/atrian/devmetrics/internal/appconfig/agentconfig"
	"github.com/atrian/devmetrics/pkg/logger"
)

type (
	// gauge основные метрики производительности
	gauge float64
	// counter счетчики опроса параметров системы
	counter int64
)

// Agent - структура
type Agent struct {
	// config конфигурация агента сбора метрик: интервалы опроса и отправки, адрес сервера, ключ для подписи метрик
	config *agentconfig.Config
	// metrics in memory хранилище для собираемых метрик
	metrics *MetricsDics
	// logger интерфейс логгера, в приложении используется ZAP логгер
	logger logger.ILogger
}

// Run запуск основных функций: сбор статистики и отправка на сервер с определенным интервалом
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
	go func() {
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
	}()

	a.RunProfiler()
}

// NewAgent подготовка зависимостей пакета: логгер, конфигурация, временное хранилище метрик
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

// RunProfiler запуск профайлера приложения на свободном порту, данные для подключения выводятся в лог
func (a *Agent) RunProfiler() {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		a.logger.Error("Can't create listener for PPROF server", err)
	}

	a.logger.Info(fmt.Sprintf("Profiler started @ %v", listener.Addr()))

	http.Serve(listener, nil)
}

// RefreshRuntimeStats обновление метрик из пакета runtime (runtime.MemStats)
func (a *Agent) RefreshRuntimeStats() {
	a.metrics.updateRuntimeMetrics()

	a.logger.Info(fmt.Sprintf("Runtime stats updated. PollCount: %v",
		int64(a.metrics.CounterDict["PollCount"].value)))
}

// RefreshGopsStats обновление метрик из пакета mem (mem.VirtualMemoryStat)
func (a *Agent) RefreshGopsStats() {
	a.metrics.updateGopsMetrics()

	a.logger.Info(fmt.Sprintf("Gops stats updated. PollCount: %v",
		int64(a.metrics.CounterDict["PollCount"].value)))
}

// UploadStats отправка метрик на сервер
func (a *Agent) UploadStats() {
	uploader := NewUploader(a.config, a.logger)
	uploader.SendAllStats(a.metrics)
	a.logger.Info("Upload stats")
}
