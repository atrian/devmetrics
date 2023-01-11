package agentconfig

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/atrian/devmetrics/pkg/logger"
)

var (
	address, hashKey *string
	reportInterval   *time.Duration
	pollInterval     *time.Duration
)

// Config конфигурация приложения отправки метрик
type Config struct {
	Agent  AgentConfig // Agent конфигурация параметров сбора и отправки
	HTTP   HTTPConfig  // HTTP конфигурация транспорта
	logger logger.ILogger
}

// AgentConfig конфигурация параметров сбора и отправки метрик
type AgentConfig struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"`   // PollInterval интервал сбора метрик, по умолчанию 2 секунды
	ReportInterval time.Duration `env:"REPORT_INTERVAL"` // ReportInterval интервал отправки метрик на сервер, по умолчанию 10 секунд
	HashKey        string        `env:"KEY"`             // HashKey ключ подписи метрик. Если пустой - метрики не подписываются
}

// HTTPConfig конфигурация транспорта
type HTTPConfig struct {
	Protocol    string // Protocol протокол передачи, по умолчанию http
	Address     string `env:"ADDRESS"` // Address адрес сервера, по умолчанию 127.0.0.1:8080
	URLTemplate string // URLTemplate шаблон, по умолчанию %v://%v/
	ContentType string // ContentType по умолчанию application/json
}

// NewConfig собирает конфигурацию из значений по умолчанию, переданных флагов и переменных окружения
// приоритет по возрастанию: умолчание > флаги > переменные среды
func NewConfig(logger logger.ILogger) *Config {
	config := Config{
		logger: logger,
	}
	config.loadAgentConfig()
	config.loadHTTPConfig()
	config.loadAgentFlags()
	config.loadAgentEnvConfiguration()
	return &config
}

// loadHTTPConfig загрузка конфигурации опроса и отправки по умолчанию
func (config *Config) loadAgentConfig() {
	config.Agent = AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}
}

// loadHTTPConfig загрузка конфигурации транспорта по умолчанию
func (config *Config) loadHTTPConfig() {
	config.HTTP = HTTPConfig{
		Protocol:    "http",
		URLTemplate: "%v://%v/",
		ContentType: "application/json",
	}
}

// loadAgentFlags загрузка конфигурации из флагов
func (config *Config) loadAgentFlags() {
	address = flag.String("a", "127.0.0.1:8080", "Address and port used for agent.")
	reportInterval = flag.Duration("r", 10*time.Second, "Metrics upload interval in seconds.")
	pollInterval = flag.Duration("p", 2*time.Second, "Metrics pool interval.")
	hashKey = flag.String("k", "", "Key for metrics sign")

	flag.Parse()

	config.HTTP.Address = *address
	config.Agent.ReportInterval = *reportInterval
	config.Agent.PollInterval = *pollInterval
	config.Agent.HashKey = *hashKey
}

// loadAgentEnvConfiguration загрузка конфигурации переменных окружения
func (config *Config) loadAgentEnvConfiguration() {
	config.logger.Info("Load env configuration")

	err := env.Parse(&config.HTTP)
	if err != nil {
		config.logger.Fatal("loadAgentEnvConfiguration env.Parse config.HTTP", err)
	}

	err = env.Parse(&config.Agent)
	if err != nil {
		config.logger.Fatal("loadAgentEnvConfiguration env.Parse config.Agent", err)
	}
}
