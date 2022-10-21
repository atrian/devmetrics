package appconfig

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"time"
)

type Config struct {
	Agent AgentConfig
	HTTP  HTTPConfig
}

type AgentConfig struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
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
	config.loadHTTPConfig()
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
}
