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

type Config struct {
	Agent  AgentConfig
	HTTP   HTTPConfig
	logger logger.Logger
}

type AgentConfig struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	HashKey        string        `env:"KEY"`
}

type HTTPConfig struct {
	Protocol    string
	Address     string `env:"ADDRESS"`
	URLTemplate string
	ContentType string
}

func NewConfig(logger logger.Logger) *Config {
	config := Config{
		logger: logger,
	}
	config.loadAgentConfig()
	config.loadHTTPConfig()
	config.loadAgentFlags()
	config.loadAgentEnvConfiguration()
	return &config
}

func (config *Config) loadAgentConfig() {
	config.Agent = AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}
}

func (config *Config) loadHTTPConfig() {
	config.HTTP = HTTPConfig{
		Protocol:    "http",
		URLTemplate: "%v://%v/",
		ContentType: "application/json",
	}
}

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
