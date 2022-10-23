package agentConfig

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
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
	address := flag.String("a", "127.0.0.1:8080", "Address and port used for agent.")
	reportInterval := flag.Duration("r", 10*time.Second, "Metrics upload interval in seconds.")
	pollInterval := flag.Duration("p", 2*time.Second, "Metrics pool interval.")

	flag.Parse()

	config.HTTP.Address = *address
	config.Agent.ReportInterval = *reportInterval
	config.Agent.PollInterval = *pollInterval
}

func (config *Config) loadAgentEnvConfiguration() {
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
