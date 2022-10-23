package appconfig

import (
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Agent  AgentConfig
	Server ServerConfig
	HTTP   HTTPConfig
}

type AgentConfig struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}

type ServerConfig struct {
	StoreInterval      time.Duration `env:"STORE_INTERVAL"`
	StoreFile          string        `env:"STORE_FILE"`
	Restore            bool          `env:"RESTORE"`
	MetricTemplateFile string
}

type HTTPConfig struct {
	Protocol    string
	Address     string `env:"ADDRESS"`
	URLTemplate string
	ContentType string
}

func NewConfig() *Config {
	config := Config{}
	config.loadDefaultConfiguration()
	config.loadEnvConfiguration()
	return &config
}

func (config *Config) loadDefaultConfiguration() {
	config.loadAgentConfig()
	config.loadServerConfig()
	config.loadHTTPConfig()
}

func (config *Config) loadAgentConfig() {
	config.Agent = AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}
}

func (config *Config) loadServerConfig() {
	config.Server = ServerConfig{
		StoreInterval:      300 * time.Second,
		StoreFile:          "/tmp/devops-metrics-db.json",
		Restore:            true,
		MetricTemplateFile: "internal/server/templates/metricTemplate.html",
	}
}

func (config *Config) loadHTTPConfig() {
	config.HTTP = HTTPConfig{
		Protocol:    "http",
		Address:     "127.0.0.1:8080",
		URLTemplate: "%v://%v/",
		ContentType: "application/json",
	}
}

func (config *Config) loadEnvConfiguration() {
	fmt.Println("Load env configuration")

	err := env.Parse(&config.HTTP)
	if err != nil {
		log.Fatal(err)
	}

	err = env.Parse(&config.Agent)
	if err != nil {
		log.Fatal(err)
	}

	err = env.Parse(&config.Server)
	if err != nil {
		log.Fatal(err)
	}
}
