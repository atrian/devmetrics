package storage

import (
	"encoding/json"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/dto"
)

type MemoryStorage struct {
	metrics     *MetricsDicts
	config      *serverconfig.Config
	logger      *zap.Logger
	silentStore bool
}

var _ Repository = (*MemoryStorage)(nil)

func NewMemoryStorage(config *serverconfig.Config, logger *zap.Logger) *MemoryStorage {
	storage := MemoryStorage{
		metrics: NewMetricsDicts(),
		config:  config,
		logger:  logger,
	}
	return &storage
}

func (s *MemoryStorage) StoreGauge(name string, value float64) error {
	s.metrics.GaugeDict[name] = gauge(value)
	if !s.silentStore {
		err := s.syncWithFileOnUpdate()
		if err != nil {
			s.logger.Error("StoreGauge syncWithFileOnUpdate", zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *MemoryStorage) GetGauge(name string) (float64, bool) {
	value, exist := s.metrics.GaugeDict[name]
	return float64(value), exist
}

func (s *MemoryStorage) StoreCounter(name string, value int64) error {
	s.metrics.CounterDict[name] += counter(value)
	if !s.silentStore {
		err := s.syncWithFileOnUpdate()
		if err != nil {
			s.logger.Error("StoreCounter syncWithFileOnUpdate", zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *MemoryStorage) GetCounter(name string) (int64, bool) {
	value, exist := s.metrics.CounterDict[name]
	return int64(value), exist
}

func (s *MemoryStorage) GetMetrics() *MetricsDicts {
	return s.metrics
}

// DumpToFile сохраняем данные из памяти в файл в json формате
func (s *MemoryStorage) DumpToFile(filename string) error {
	s.logger.Debug("Dump data to file")

	// STORE_FILE - пустое значение — отключает функцию записи на диск
	if filename == "" {
		s.logger.Debug("Filename is empty, abort dumping")
		return nil
	}

	metricWriter, err := NewMetricWriter(filename)
	if err != nil {
		s.logger.Error("NewMetricWriter error", zap.Error(err))
	}

	defer metricWriter.Close()

	metricsDTO := make([]dto.Metrics, 0, len(s.metrics.GaugeDict)+len(s.metrics.GaugeDict))

	// собираем gauge метрики в общий слайс с метриками
	for key, metric := range s.metrics.GaugeDict {
		floatVal := float64(metric)
		metricDTO := dto.Metrics{
			ID:    key,
			MType: "gauge",
			Delta: nil,
			Value: &floatVal,
		}
		metricsDTO = append(metricsDTO, metricDTO)
	}

	// собираем counter метрики в общий слайс с метриками
	for key, metric := range s.metrics.CounterDict {
		intVal := int64(metric)
		metricDTO := dto.Metrics{
			ID:    key,
			MType: "counter",
			Delta: &intVal,
			Value: nil,
		}
		metricsDTO = append(metricsDTO, metricDTO)
	}

	// пишем все метрики в JSON
	if err := metricWriter.WriteMetric(&metricsDTO); err != nil {
		s.logger.Error("WriteMetric error", zap.Error(err))
		return err
	}

	return nil
}

// RestoreFromFile Восстановление данных из файла
func (s *MemoryStorage) RestoreFromFile(filename string) error {
	s.logger.Info("Restore metrics from file")

	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		s.logger.Warn("RestoreFromFile can't load file", zap.Error(err))
	}

	var metrics []dto.Metrics
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&metrics)
	if err != nil {
		s.logger.Warn("Can't Decode metrics", zap.Error(err))
	}

	s.SetMetrics(metrics)

	return nil
}

func (s *MemoryStorage) SetMetrics(metrics []dto.Metrics) {
	s.silentStore = true
	for key, metricCandidate := range metrics {
		_ = key
		switch metricCandidate.MType {
		case "gauge":
			s.StoreGauge(metricCandidate.ID, *metricCandidate.Value)
		case "counter":
			s.StoreCounter(metricCandidate.ID, *metricCandidate.Delta)
		default:
		}
	}
	s.silentStore = false
}

// syncWithFileOnUpdate сохраняем дамп метрик в файл при обновлении любой метрики если StoreInterval = 0
func (s *MemoryStorage) syncWithFileOnUpdate() error {
	if s.config.Server.StoreInterval == 0 {
		err := s.DumpToFile(s.config.Server.StoreFile)
		if err != nil {
			s.logger.Error("syncWithFileOnUpdate", zap.Error(err))
			return err
		}
	}
	return nil
}

// RunOnStart метод вызывается при старте хранилища
func (s *MemoryStorage) RunOnStart() {
	// RESTORE (по умолчанию true) — булево значение (true/false), определяющее,
	// загружать или нет начальные значения из указанного файла при старте сервера.
	if s.config.Server.Restore {
		err := s.RestoreFromFile(s.config.Server.StoreFile)
		if err != nil {
			s.logger.Error("RunOnStart - RestoreFromFile call", zap.Error(err))
		}
	}

	// STORE_INTERVAL (по умолчанию 300) — интервал времени в секундах,
	// по истечении которого текущие показания сервера сбрасываются на диск
	// (значение 0 — делает запись синхронной).
	if s.config.Server.StoreInterval != 0 {
		s.runMetricsDumpTicker()
	}
}

// RunOnClose метод вызывается при штатном завершении
func (s *MemoryStorage) RunOnClose() {
	s.logger.Info("Dump metrics to file before shutdown")
	err := s.DumpToFile(s.config.Server.StoreFile)
	if err != nil {
		s.logger.Error("RunOnClose DumpToFile", zap.Error(err))
	}
}

// runMetricsDumpTicker дамп хранилища из памяти в файл с запуском по тикеру
func (s *MemoryStorage) runMetricsDumpTicker() {
	// запускаем тикер дампа статистики
	dumpMetricsTicker := time.NewTicker(s.config.Server.StoreInterval)

	s.logger.Info("Run metrics dump", zap.Duration("StoreInterval", s.config.Server.StoreInterval))

	go func() {
		for dumpTime := range dumpMetricsTicker.C {
			err := s.DumpToFile(s.config.Server.StoreFile)
			if err != nil {
				s.logger.Error("DumpToFile go func", zap.Error(err))
			}

			s.logger.Info("Metrics dump time", zap.Time("dumpTime", dumpTime))
		}
	}()
}
