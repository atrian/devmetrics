package appconfig

import (
	"flag"
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

func NewServerConfig() *Config {
	config := Config{}
	config.loadServerConfig()
	config.loadHTTPConfig()
	config.loadServerFlags()
	config.loadServerEnvConfiguration()
	return &config
}

func (config *Config) loadServerConfig() {
	config.Server = ServerConfig{
		StoreInterval:      300 * time.Second,
		StoreFile:          "/tmp/devops-metrics-db.json",
		Restore:            true,
		MetricTemplateFile: "internal/server/templates/metricTemplate.html",
	}
}

func NewAgentConfig() *Config {
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
		Address:     "127.0.0.1:8080",
		URLTemplate: "%v://%v/",
		ContentType: "application/json",
	}
}

func (config *Config) loadServerEnvConfiguration() {
	fmt.Println("Load env configuration")

	err := env.Parse(&config.HTTP)
	if err != nil {
		log.Fatal(err)
	}

	err = env.Parse(&config.Server)
	if err != nil {
		log.Fatal(err)
	}
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

func (config *Config) loadServerFlags() {
	address := flag.String("a", "127.0.0.1:8080", "Address and port used for server and agent.")
	file := flag.String("f", "/tmp/devops-metrics-db.json", "Where to store metrics dump file.")
	restore := flag.Bool("r", true, "Restore metrics from dump file on server start.")
	storeInterval := flag.Int64("i", 300, "Metrics dump interval in seconds.")

	flag.Parse()

	config.HTTP.Address = *address
	config.Server.StoreFile = *file
	config.Server.Restore = *restore
	config.Server.StoreInterval = time.Duration(*storeInterval) * time.Second
}

func (config *Config) loadAgentFlags() {
	address := flag.String("a", "127.0.0.1:8080", "Address and port used for agent.")
	reportInterval := flag.Int64("r", 10, "Metrics upload interval in seconds.")
	pollInterval := flag.Int64("p", 2, "Metrics pool interval in seconds.")

	flag.Parse()

	config.HTTP.Address = *address
	config.Agent.ReportInterval = time.Duration(*reportInterval) * time.Second
	config.Agent.PollInterval = time.Duration(*pollInterval) * time.Second
}
