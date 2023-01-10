package serverconfig

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/atrian/devmetrics/pkg/logger"
)

var (
	address       *string
	file          *string
	hashKey       *string
	dsn           *string
	restore       *bool
	storeInterval *time.Duration
	profile       *bool
)

type Config struct {
	Server ServerConfig
	HTTP   HTTPConfig
	logger logger.Logger
}

type ServerConfig struct {
	StoreInterval      time.Duration `env:"STORE_INTERVAL"`
	StoreFile          string        `env:"STORE_FILE"`
	Restore            bool          `env:"RESTORE"`
	MetricTemplateFile string
	HashKey            string `env:"KEY"`
	DBDSN              string `env:"DATABASE_DSN"`
	ProfileApp         bool
}

type HTTPConfig struct {
	Protocol    string
	Address     string `env:"ADDRESS"`
	URLTemplate string
	ContentType string
}

func NewServerConfig(logger logger.Logger) *Config {
	conf := Config{
		logger: logger,
	}
	conf.loadServerConfig()
	conf.loadHTTPConfig()
	conf.loadServerFlags()
	conf.loadServerEnvConfiguration()
	return &conf
}

func (config *Config) loadServerConfig() {
	config.Server = ServerConfig{
		StoreInterval:      300 * time.Second,
		StoreFile:          "tmp/devops-metrics-db.json",
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
		config.logger.Fatal("loadServerEnvConfiguration env.Parse config.HTTP", err)
	}

	err = env.Parse(&config.Server)
	if err != nil {
		config.logger.Fatal("loadServerEnvConfiguration env.Parse config.Server", err)
	}
}

func (config *Config) loadServerFlags() {
	address = flag.String("a", "127.0.0.1:8080", "Address and port used for server and agent.")
	file = flag.String("f", "tmp/devops-metrics-db.json", "Where to store metrics dump file.")
	restore = flag.Bool("r", true, "Restore metrics from dump file on server start.")
	storeInterval = flag.Duration("i", 300*time.Second, "Metrics dump interval in seconds.")
	hashKey = flag.String("k", "", "Key for metrics sign validation")
	dsn = flag.String("d", "", "DSN for PostgreSQL server")
	profile = flag.Bool("p", false, "Enable pprof profiler")

	flag.Parse()

	config.HTTP.Address = *address
	config.Server.StoreFile = *file
	config.Server.Restore = *restore
	config.Server.StoreInterval = *storeInterval
	config.Server.HashKey = *hashKey
	config.Server.DBDSN = *dsn
	config.Server.ProfileApp = *profile
}
