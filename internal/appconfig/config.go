package appconfig

import (
	"time"
)

type Config struct {
	Agent AgentConfig
	HTTP  HTTPConfig
}

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type HTTPConfig struct {
	Protocol    string
	Server      string
	Port        uint
	URLTemplate string
	ContentType string
}

func NewConfig() *Config {
	config := Config{}
	config.loadDefaultConfiguration()
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
		Server:      "127.0.0.1",
		Port:        8080,
		URLTemplate: "%v://%v:%d/update", // <ПРОТОКОЛ>://<АДРЕС_СЕРВЕРА>/update
		ContentType: "application/json",
	}
}
