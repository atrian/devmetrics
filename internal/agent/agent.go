// Package agent - клиентская часть приложения по сбору метрик.
// собирает фиксированный список метрик и отправляет на сервер
// Интервалы сбора метрик и отправки настраиваются.
// Данные отправляются в формате JSON в пакетном режиме, применяется Gzip сжатие.
// В приложении доступен профилировщик
package agent

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
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

// Agent - основное приложение агента сборщика
type Agent struct {
	// config конфигурация агента сбора метрик: интервалы опроса и отправки, адрес сервера, ключ для подписи метрик
	config *agentconfig.Config
	// metrics in memory хранилище для собираемых метрик
	metrics *MetricsDics
	// uploader для загрузки метрик на сервер
	uploader *Uploader
	// logger интерфейс логгера, в приложении используется ZAP логгер
	logger logger.Logger
	// profiler сервер профилировщика
	profiler http.Server
}

// Run запуск основных функций: сбор статистики и отправка на сервер с определенным интервалом
func (a *Agent) Run(ctx context.Context) {
	graceShutdown := make(chan struct{})

	a.logger.Info(
		fmt.Sprintf("Agent started. PollInterval: %v, ReportInterval: %v, Server address: %v",
			a.config.Agent.PollInterval,
			a.config.Agent.ReportInterval,
			a.config.Transport.AddressHTTP))

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
			case <-ctx.Done():
				// при завершении контекста выполняем последнюю отправку метрик
				// закрываем сервер профилировщика, дожидаемся завершения операции и выходим из приложения

				go func() {
					a.Stop(graceShutdown)
				}()

				a.logger.Info("Agent shutdown gracefully")

				return
			}
		}
	}()

	a.RunProfiler()
	<-graceShutdown
}

// NewAgent подготовка зависимостей пакета: логгер, конфигурация, временное хранилище метрик
func NewAgent() *Agent {
	// подключаем логгер
	agentLogger := logger.NewZapLogger()
	defer agentLogger.Sync()

	config := agentconfig.NewConfig(agentLogger)

	agent := &Agent{
		config:   config,
		metrics:  NewMetricsDicts(agentLogger),
		uploader: NewUploader(config, agentLogger),
		logger:   agentLogger,
	}

	err := agent.RefreshAgentIp()
	if err != nil {
		agent.logger.Error("Can't get agent IP address", err)
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

	if profilerErr := a.profiler.Serve(listener); profilerErr != http.ErrServerClosed {
		// ошибки старта или остановки Listener
		a.logger.Error("Profiler server ListenAndServe: %v", profilerErr)
	}
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
	a.uploader.SendAllStats(a.metrics)
	a.logger.Info("Upload stats")
}

// Stop операции при завершении приложения
func (a *Agent) Stop(grace chan struct{}) {
	defer close(grace)

	// Отправляем все текущие метрики
	a.UploadStats()
	a.logger.Info("Last metrics sent")

	// Завершаем сервер профилирования
	if err := a.profiler.Shutdown(context.Background()); err != nil {
		// ошибки закрытия Listener
		a.logger.Error("Profiler server Shutdown err", err)
	}

	// закрываем GRPC соединение
	if a.uploader.GRPCConnection != nil {
		err := a.uploader.GRPCConnection.Close()
		if err != nil {
			a.logger.Error("Error on GRPC connection close", err)
		}
	}

	a.logger.Info("Profiler closed")
}

// RefreshAgentIp обновляет IP адрес агента в загруженной конфигурации
func (a *Agent) RefreshAgentIp() error {
	host, err := os.Hostname()
	if err != nil {
		return err
	}

	IPs, err := net.LookupHost(host)
	if err != nil {
		return err
	}

	a.config.Agent.AgentIP = net.ParseIP(IPs[0])
	a.logger.Info(fmt.Sprintf("Agent IP loaded: %v", a.config.Agent.AgentIP.String()))
	return nil
}
