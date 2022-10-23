package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/atrian/devmetrics/internal/appconfig"
	"github.com/atrian/devmetrics/internal/dto"
)

type MemoryStorage struct {
	metrics *MetricsDicts
	config  *appconfig.Config
}

func NewMemoryStorage(config *appconfig.Config) *MemoryStorage {
	storage := MemoryStorage{
		metrics: NewMetricsDicts(),
		config:  config,
	}
	return &storage
}

func (s MemoryStorage) StoreGauge(name string, value float64) {
	s.metrics.GaugeDict[name] = gauge(value)
	s.syncWithFileOnUpdate()
}

func (s MemoryStorage) GetGauge(name string) (float64, bool) {
	value, exist := s.metrics.GaugeDict[name]
	return float64(value), exist
}

func (s MemoryStorage) StoreCounter(name string, value int64) {
	s.metrics.CounterDict[name] += counter(value)
	s.syncWithFileOnUpdate()
}

func (s MemoryStorage) GetCounter(name string) (int64, bool) {
	value, exist := s.metrics.CounterDict[name]
	return int64(value), exist
}

func (s MemoryStorage) GetMetrics() *MetricsDicts {
	return s.metrics
}

func (s MemoryStorage) DumpToFile(filename string) error {
	fmt.Println("Dump data to file")

	// STORE_FILE - пустое значение — отключает функцию записи на диск
	if filename == "" {
		return nil
	}

	metricWriter, err := NewMetricWriter(filename)
	if err != nil {
		log.Fatal(err)
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
		return err
	}

	return nil
}

// RestoreFromFile Восстановление данных из файла
func (s MemoryStorage) RestoreFromFile(filename string) error {
	fmt.Println("Restore metrics from file")

	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println("Can't load file:", err)
	}

	var metrics []dto.Metrics
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&metrics)
	if err != nil {
		fmt.Println("Can't Decode metrics:", err)
	}

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

	return nil
}

// syncWithFileOnUpdate сохраняем дамп метрик в файл при обновлении любой метрики если StoreInterval = 0
func (s MemoryStorage) syncWithFileOnUpdate() {
	if s.config.Server.StoreInterval == 0 {
		err := s.DumpToFile(s.config.Server.StoreFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}
