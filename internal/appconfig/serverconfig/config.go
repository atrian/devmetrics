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

// Config конфигурация сервера приема метрик
type Config struct {
	Server ServerConfig
	HTTP   HTTPConfig
	logger logger.ILogger
}

// ServerConfig основная конфигурация сервера для хранения метрик
type ServerConfig struct {
	StoreInterval      time.Duration `env:"STORE_INTERVAL"` // StoreInterval интервал сохранения накопленных метрик в файл на диске, по умолчанию раз в 5 минут
	StoreFile          string        `env:"STORE_FILE"`     // StoreFile файл для сохранения накопленных метрик на диске
	Restore            bool          `env:"RESTORE"`        // Restore флаг периодического сброса накопленных метрик в файл на диск
	MetricTemplateFile string        // MetricTemplateFile шаблон вывода метрик в HTML формате
	HashKey            string        `env:"KEY"`          // HashKey ключ для проверки подписи метрик
	DBDSN              string        `env:"DATABASE_DSN"` // DBDSN строка соединения с базой данных (PGSQL)
	ProfileApp         bool          // ProfileApp флаг разрешающий маршруты для просмотра профиля pprof приложения
}

// HTTPConfig конфигурация старта веб сервера для приема запросов
type HTTPConfig struct {
	Address     string `env:"ADDRESS"` // Address адрес сервера
	ContentType string // ContentType устанавливается в заголовках ответа
}

// NewServerConfig собирает конфигурацию из значений по умолчанию, переданных флагов и переменных окружения
// приоритет по возрастанию: умолчание > флаги > переменные среды
func NewServerConfig(logger logger.ILogger) *Config {
	conf := Config{
		logger: logger,
	}
	conf.loadServerConfig()
	conf.loadHTTPConfig()
	conf.loadServerFlags()
	conf.loadServerEnvConfiguration()
	return &conf
}

// NewServerConfigWithoutFlags собирает конфигурацию из значений по умолчанию и переменных окружения
// приоритет по возрастанию: умолчание > переменные среды
// используется для тестирования
func NewServerConfigWithoutFlags(logger logger.ILogger) *Config {
	conf := Config{
		logger: logger,
	}
	conf.loadServerConfig()
	conf.loadHTTPConfig()
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
		Address:     "127.0.0.1:8080",
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
