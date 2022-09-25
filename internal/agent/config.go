package agent

import (
	"time"
)

type Config struct {
	Agent AgentConfig
	Http  HttpConfig
}

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type HttpConfig struct {
	Server      string
	Port        uint
	UrlTemplate string
	ContentType string
}

func NewConfig() *Config {
	config := Config{}
	config.loadDefaultConfiguration()
	return &config
}

func (config *Config) loadDefaultConfiguration() {
	config.loadAgentConfig()
	config.loadHttpConfig()
}

func (config *Config) loadAgentConfig() {
	config.Agent = AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}
}

func (config *Config) loadHttpConfig() {
	config.Http = HttpConfig{
		Server:      "127.0.0.1",
		Port:        8080,
		UrlTemplate: "http://%v:%d/update/%v/%v/%v", // http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		ContentType: "Content-Type: text/plain",
	}
}
