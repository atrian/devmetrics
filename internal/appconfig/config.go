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
		Server:      "127.0.0.1",
		Port:        8080,
		URLTemplate: "http://%v:%d/update/%v/%v/%v", // http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		ContentType: "text/plain",
	}
}
