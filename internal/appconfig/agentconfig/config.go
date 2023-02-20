// Package agentconfig Конфигурация агента сборщика метрик.
// Содержит интервалы сборки и отправки, адрес сервера, ключ для цифровой подписи метрик
package agentconfig

import (
	"encoding/json"
	"flag"
	"net"
	"os"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/atrian/devmetrics/pkg/logger"
)

var (
	address, addressGrpc, hashKey, cryptoKey, jsonConf *string
	reportInterval                                     *time.Duration
	pollInterval                                       *time.Duration
)

// Config конфигурация приложения отправки метрик
type Config struct {
	Transport TransportConfig // конфигурация транспорта
	logger    logger.Logger
	Agent     AgentConfig // Agent конфигурация параметров сбора и отправки
}

// ConfDummy шаблон для парсинга JSON конфигурации
type ConfDummy struct {
	Address        string `json:"address,omitempty"`
	AddressGRPC    string `json:"address_grpc,omitempty"`
	ReportInterval string `json:"report_interval,omitempty"`
	PollInterval   string `json:"poll_interval,omitempty"`
	CryptoKey      string `json:"crypto_key,omitempty"`
}

// AgentConfig конфигурация параметров сбора и отправки метрик
type AgentConfig struct {
	AgentIP        net.IP        // AgentIP адрес агента. Определяется при старте
	CryptoKey      string        `env:"CRYPTO_KEY"`      // CryptoKey путь до файла с публичным ключом
	HashKey        string        `env:"KEY"`             // HashKey ключ подписи метрик. Если пустой - метрики не подписываются
	PollInterval   time.Duration `env:"POLL_INTERVAL"`   // PollInterval интервал сбора метрик, по умолчанию 2 секунды
	ReportInterval time.Duration `env:"REPORT_INTERVAL"` // ReportInterval интервал отправки метрик на сервер, по умолчанию 10 секунд
}

// TransportConfig конфигурация транспорта
type TransportConfig struct {
	Protocol    string // Protocol протокол передачи, по умолчанию http
	AddressHTTP string `env:"ADDRESS"`      // AddressHTTP адрес WEB сервера, по умолчанию 127.0.0.1:8080
	AddressGRPC string `env:"ADDRESS_GRPC"` // AddressGRPC адрес GRPC сервера
	URLTemplate string // URLTemplate шаблон, по умолчанию %v://%v/
	ContentType string // ContentType по умолчанию application/json
}

// NewConfig собирает конфигурацию из значений по умолчанию, json файла конфигурации, переданных флагов и переменных окружения
// приоритет по возрастанию: умолчание > json файл конфигурации > флаги > переменные среды
func NewConfig(logger logger.Logger) *Config {
	config := Config{
		logger: logger,
	}

	// конфигурация по умолчанию
	config.loadAgentConfig()
	config.loadHTTPConfig()

	config.parseFlags()
	config.loadJSONConfiguration()
	config.loadAgentFlags()
	config.loadAgentEnvConfiguration()
	config.selectProtocol() // Если передан адрес GRPC используем его в качестве транспорта
	return &config
}

// loadHTTPConfig загрузка конфигурации опроса и отправки по умолчанию
func (config *Config) loadAgentConfig() {
	config.Agent = AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}
}

// loadHTTPConfig загрузка конфигурации транспорта по умолчанию
func (config *Config) loadHTTPConfig() {
	config.Transport = TransportConfig{
		Protocol:    "http",
		AddressHTTP: "127.0.0.1:8080",
		URLTemplate: "%v://%v/",
		ContentType: "application/json",
	}
}

// parseFlags парсит все флаги приложения
func (config *Config) parseFlags() {
	jsonConf = flag.String("config", "", "Path to JSON configuration file")
	flag.StringVar(jsonConf, "c", *jsonConf, "alias for -config")

	address = flag.String("a", "127.0.0.1:8080", "AddressHTTP and port used for agent.")
	addressGrpc = flag.String("ag", "127.0.0.1:9876", "AddressHTTP and port used for GRPC connection.")
	reportInterval = flag.Duration("r", 10*time.Second, "Metrics upload interval in seconds.")
	pollInterval = flag.Duration("p", 2*time.Second, "Metrics pool interval.")
	hashKey = flag.String("k", "", "Key for metrics sign")
	cryptoKey = flag.String("crypto-key", "", "Path to public PEM key")

	flag.Parse()
}

// loadAgentFlags загрузка конфигурации из флагов
func (config *Config) loadAgentFlags() {
	if isFlagPassed("a") {
		config.Transport.AddressHTTP = *address
	}

	if isFlagPassed("ag") {
		config.Transport.AddressGRPC = *addressGrpc
	}

	if isFlagPassed("r") {
		config.Agent.ReportInterval = *reportInterval
	}

	if isFlagPassed("p") {
		config.Agent.PollInterval = *pollInterval
	}

	if isFlagPassed("k") {
		config.Agent.HashKey = *hashKey
	}

	if isFlagPassed("crypto-key") {
		config.Agent.CryptoKey = *cryptoKey
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

	config.Transport.AddressHTTP = dummy.Address
	config.Transport.AddressGRPC = dummy.AddressGRPC
	config.Agent.CryptoKey = dummy.CryptoKey

	parsedReportInterval, _ := time.ParseDuration(dummy.ReportInterval)
	config.Agent.ReportInterval = parsedReportInterval

	parsedPoolInterval, _ := time.ParseDuration(dummy.PollInterval)
	config.Agent.PollInterval = parsedPoolInterval

	config.logger.Info("JSON configuration loaded")
}

// loadAgentEnvConfiguration загрузка конфигурации переменных окружения
func (config *Config) loadAgentEnvConfiguration() {
	config.logger.Info("Load env configuration")

	err := env.Parse(&config.Transport)
	if err != nil {
		config.logger.Fatal("loadAgentEnvConfiguration env.Parse config.Transport", err)
	}

	err = env.Parse(&config.Agent)
	if err != nil {
		config.logger.Fatal("loadAgentEnvConfiguration env.Parse config.Agent", err)
	}
}

func (config *Config) selectProtocol() {
	if config.Transport.AddressGRPC != "" {
		config.Transport.Protocol = "grpc"
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
