// Package serverconfig Конфигурация сервера хранения метрик.
// Содержит адрес сервера, dsn соединения с БД, ключ для проверки цифровых подписей метрик, флаг профилирования
package serverconfig

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/atrian/devmetrics/pkg/logger"
)

var (
	address       *string
	file          *string
	hashKey       *string
	cryptoKey     *string
	dsn           *string
	jsonConf      *string
	trustedSubnet *string
	storeInterval *time.Duration
	restore       *bool
	profile       *bool
)

// Config конфигурация сервера приема метрик
type Config struct {
	HTTP   HTTPConfig
	logger logger.Logger
	Server ServerConfig
}

// ConfDummy шаблон для парсинга JSON конфигурации
type ConfDummy struct {
	Address       string `json:"address,omitempty"`
	StoreInterval string `json:"store_interval,omitempty"`
	StoreFile     string `json:"store_file,omitempty"`
	DatabaseDsn   string `json:"database_dsn,omitempty"`
	CryptoKey     string `json:"crypto_key,omitempty"`
	TrustedSubnet string `json:"trusted_subnet,omitempty"`
	Restore       bool   `json:"restore,omitempty"`
}

// ServerConfig основная конфигурация сервера для хранения метрик
type ServerConfig struct {
	StoreFile          string        `env:"STORE_FILE"` // StoreFile файл для сохранения накопленных метрик на диске
	MetricTemplateFile string        // MetricTemplateFile шаблон вывода метрик в HTML формате
	HashKey            string        `env:"KEY"`            // HashKey ключ для проверки подписи метрик
	CryptoKey          string        `env:"CRYPTO_KEY"`     // CryptoKey путь до файла с приватным ключом
	DBDSN              string        `env:"DATABASE_DSN"`   // DBDSN строка соединения с базой данных (PGSQL)
	TrustedSubnet      string        `env:"TRUSTED_SUBNET"` // TrustedSubnet Получение метрик только из доверенной сети. Принимает строковое представление бесклассовой адресации (CIDR)
	StoreInterval      time.Duration `env:"STORE_INTERVAL"` // StoreInterval интервал сохранения накопленных метрик в файл на диске, по умолчанию раз в 5 минут
	Restore            bool          `env:"RESTORE"`        // Restore флаг периодического сброса накопленных метрик в файл на диск
	ProfileApp         bool          // ProfileApp флаг разрешающий маршруты для просмотра профиля pprof приложения
}

// HTTPConfig конфигурация старта веб сервера для приема запросов
type HTTPConfig struct {
	Address     string `env:"ADDRESS"` // Address адрес сервера
	ContentType string // ContentType устанавливается в заголовках ответа
}

// NewServerConfig собирает конфигурацию из значений по умолчанию, переданных флагов и переменных окружения
// приоритет по возрастанию: умолчание > флаги > переменные среды
func NewServerConfig(logger logger.Logger) *Config {
	conf := Config{
		logger: logger,
	}
	conf.loadServerConfig()
	conf.loadHTTPConfig()
	conf.parseFlags()
	conf.loadJSONConfiguration()
	conf.loadServerFlags()
	conf.loadServerEnvConfiguration()
	return &conf
}

// NewServerConfigWithoutFlags собирает конфигурацию из значений по умолчанию и переменных окружения
// приоритет по возрастанию: умолчание > переменные среды
// используется для тестирования
// TODO - удалить метод, переписать тесты
func NewServerConfigWithoutFlags(logger logger.Logger) *Config {
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
		StoreFile:          "tpm",
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

// loadJSONConfiguration извлекает путь к JSON конфигу из флагов -c -config или переменной окружения CONFIG
// Открывает файл и загружает конфигурацию. У JSON конфигурации самый низкий приоритет
// конфигурация может быть в дальнейшем перезаписана данными из флагов и переменных окружения
// Ошибки открытия файла или парсинга конфигурации приведут к остановке программы
func (config *Config) loadJSONConfiguration() {
	var (
		JSONConfigPath string
		dummy          ConfDummy
	)

	if *jsonConf != "" {
		JSONConfigPath = *jsonConf
	}

	if envPath, ok := os.LookupEnv("CONFIG"); ok {
		JSONConfigPath = envPath
	}

	// если путь к файлу не предоставлен, завершаем работу метода
	if JSONConfigPath == "" {
		return
	}

	cFile, err := os.Open(JSONConfigPath)
	if err != nil {
		config.logger.Fatal("loadJSONConfiguration os.Open error", err)
	}

	defer func(cFile *os.File) {
		cErr := cFile.Close()
		if cErr != nil {
			config.logger.Fatal("loadJSONConfiguration cFile.Close error", cErr)
		}
	}(cFile)

	d := json.NewDecoder(cFile)
	dErr := d.Decode(&dummy)
	if dErr != nil {
		config.logger.Fatal("loadJSONConfiguration json.Decode error", err)
	}

	config.HTTP.Address = dummy.Address
	config.Server.Restore = dummy.Restore
	config.Server.StoreFile = dummy.StoreFile
	config.Server.DBDSN = dummy.DatabaseDsn
	config.Server.TrustedSubnet = dummy.TrustedSubnet
	config.Server.CryptoKey = dummy.CryptoKey

	parsedStoreInterval, _ := time.ParseDuration(dummy.StoreInterval)
	config.Server.StoreInterval = parsedStoreInterval

	config.logger.Info("JSON configuration loaded")
}

// loadServerEnvConfiguration загрузка конфигурации переменных окружения
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

// parseFlags парсит все флаги приложения
func (config *Config) parseFlags() {
	jsonConf = flag.String("config", "", "Path to JSON configuration file")
	flag.StringVar(jsonConf, "c", *jsonConf, "alias for -config")

	address = flag.String("a", config.HTTP.Address, "Address and port used for server and agent.")
	file = flag.String("f", config.Server.StoreFile, "Where to store metrics dump file.")
	restore = flag.Bool("r", config.Server.Restore, "Restore metrics from dump file on server start.")
	storeInterval = flag.Duration("i", config.Server.StoreInterval, "Metrics dump interval in seconds.")
	hashKey = flag.String("k", "", "Key for metrics sign validation")
	cryptoKey = flag.String("crypto-key", config.Server.CryptoKey, "Path to private PEM key")
	dsn = flag.String("d", config.Server.DBDSN, "DSN for PostgreSQL server")
	trustedSubnet = flag.String("t", config.Server.TrustedSubnet, "Accept metrics only from trusted network. CIDR.")
	profile = flag.Bool("p", false, "Enable pprof profiler")

	flag.Parse()
}

// loadServerFlags загрузка конфигурации из флагов
func (config *Config) loadServerFlags() {
	if isFlagPassed("a") {
		config.HTTP.Address = *address
	}

	if isFlagPassed("f") {
		config.Server.StoreFile = *file
	}

	if isFlagPassed("r") {
		config.Server.Restore = *restore
	}

	if isFlagPassed("i") {
		config.Server.StoreInterval = *storeInterval
	}

	if isFlagPassed("k") {
		config.Server.HashKey = *hashKey
	}

	if isFlagPassed("crypto-key") {
		config.Server.CryptoKey = *cryptoKey
	}

	if isFlagPassed("d") {
		config.Server.DBDSN = *dsn
	}

	if isFlagPassed("t") {
		config.Server.TrustedSubnet = *trustedSubnet
	}

	if isFlagPassed("p") {
		config.Server.ProfileApp = *profile
	}
}

// isFlagPassed проверка указан ли флан при запуске программы
func isFlagPassed(name string) bool {
	var found bool
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
