package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/dto"
	"github.com/atrian/devmetrics/pkg/logger"
)

type MemoryStorage struct {
	metrics     *MetricsDicts
	config      *serverconfig.Config
	logger      logger.Logger
	silentStore bool
}

var _ Repository = (*MemoryStorage)(nil)

func NewMemoryStorage(config *serverconfig.Config, logger logger.Logger) *MemoryStorage {
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
			s.logger.Error("StoreGauge syncWithFileOnUpdate", err)
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
			s.logger.Error("StoreCounter syncWithFileOnUpdate", err)
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
		s.logger.Error("NewMetricWriter error", err)
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
		s.logger.Error("WriteMetric error", err)
		return err
	}

	return nil
}

// RestoreFromFile Восстановление данных из файла
func (s *MemoryStorage) RestoreFromFile(filename string) error {
	s.logger.Info("Restore metrics from file")

	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		s.logger.Warning("RestoreFromFile can't load file")
		return err
	}

	var metrics []dto.Metrics
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&metrics)
	if err != nil {
		s.logger.Warning("Can't Decode metrics")
		return err
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
			err := s.StoreGauge(metricCandidate.ID, *metricCandidate.Value)
			if err != nil {
				s.logger.Error("SetMetrics StoreGauge error", err)
			}
		case "counter":
			err := s.StoreCounter(metricCandidate.ID, *metricCandidate.Delta)
			if err != nil {
				s.logger.Error("SetMetrics StoreCounter error", err)
			}
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
			s.logger.Error("syncWithFileOnUpdate", err)
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
			s.logger.Error("RunOnStart - RestoreFromFile call", err)
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
		s.logger.Error("RunOnClose DumpToFile", err)
	}
}

// runMetricsDumpTicker дамп хранилища из памяти в файл с запуском по тикеру
func (s *MemoryStorage) runMetricsDumpTicker() {
	// запускаем тикер дампа статистики
	dumpMetricsTicker := time.NewTicker(s.config.Server.StoreInterval)

	s.logger.Info(fmt.Sprintf("Run metrics dump. StoreInterval: %v", s.config.Server.StoreInterval))

	go func() {
		for dumpTime := range dumpMetricsTicker.C {
			err := s.DumpToFile(s.config.Server.StoreFile)
			if err != nil {
				s.logger.Error("DumpToFile go func", err)
			}

			s.logger.Info(fmt.Sprintf("Metrics dump time. dumpTime: %v", dumpTime))
		}
	}()
}
