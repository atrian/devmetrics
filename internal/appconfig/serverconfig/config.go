package serverconfig

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

var (
	address       *string
	file          *string
	hashKey       *string
	restore       *bool
	storeInterval *time.Duration
)

type Config struct {
	Server ServerConfig
	HTTP   HTTPConfig
}

type ServerConfig struct {
	StoreInterval      time.Duration `env:"STORE_INTERVAL"`
	StoreFile          string        `env:"STORE_FILE"`
	Restore            bool          `env:"RESTORE"`
	MetricTemplateFile string
	HashKey            string `env:"KEY"`
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

func (config *Config) loadServerFlags() {
	address = flag.String("a", "127.0.0.1:8080", "Address and port used for server and agent.")
	file = flag.String("f", "/tmp/devops-metrics-db.json", "Where to store metrics dump file.")
	restore = flag.Bool("r", true, "Restore metrics from dump file on server start.")
	storeInterval = flag.Duration("i", 300*time.Second, "Metrics dump interval in seconds.")
	hashKey = flag.String("k", "", "Key for metrics sign validation")

	flag.Parse()

	config.HTTP.Address = *address
	config.Server.StoreFile = *file
	config.Server.Restore = *restore
	config.Server.StoreInterval = *storeInterval
	config.Server.HashKey = *hashKey
}
