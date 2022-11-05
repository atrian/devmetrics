package serverconfig

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

var (
	address       *string
	file          *string
	hashKey       *string
	dsn           *string
	restore       *bool
	storeInterval *time.Duration
)

type Config struct {
	Server ServerConfig
	HTTP   HTTPConfig
	logger *zap.Logger
}

type ServerConfig struct {
	StoreInterval      time.Duration `env:"STORE_INTERVAL"`
	StoreFile          string        `env:"STORE_FILE"`
	Restore            bool          `env:"RESTORE"`
	MetricTemplateFile string
	HashKey            string `env:"KEY"`
	DBDSN              string `env:"DATABASE_DSN"`
}

type HTTPConfig struct {
	Protocol    string
	Address     string `env:"ADDRESS"`
	URLTemplate string
	ContentType string
}

func NewServerConfig(logger *zap.Logger) *Config {
	config := Config{
		logger: logger,
	}
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
	config.logger.Info("Load env configuration")

	err := env.Parse(&config.HTTP)
	if err != nil {
		config.logger.Fatal("loadServerEnvConfiguration env.Parse config.HTTP", zap.Error(err))
	}

	err = env.Parse(&config.Server)
	if err != nil {
		config.logger.Fatal("loadServerEnvConfiguration env.Parse config.Server", zap.Error(err))
	}
}

func (config *Config) loadServerFlags() {
	address = flag.String("a", "127.0.0.1:8080", "Address and port used for server and agent.")
	file = flag.String("f", "/tmp/devops-metrics-db.json", "Where to store metrics dump file.")
	restore = flag.Bool("r", true, "Restore metrics from dump file on server start.")
	storeInterval = flag.Duration("i", 300*time.Second, "Metrics dump interval in seconds.")
	hashKey = flag.String("k", "", "Key for metrics sign validation")
	dsn = flag.String("d", "", "DSN for PostgreSQL server")

	flag.Parse()

	config.HTTP.Address = *address
	config.Server.StoreFile = *file
	config.Server.Restore = *restore
	config.Server.StoreInterval = *storeInterval
	config.Server.HashKey = *hashKey
	config.Server.DBDSN = *dsn
}
