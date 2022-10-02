package storage

import "strconv"

type MemoryStorage struct {
	metrics *MetricsDics
}

func NewMemoryStorage() *MemoryStorage {
	storage := MemoryStorage{metrics: NewMetricsDicts()}
	return &storage
}

func (s MemoryStorage) StoreGauge(name string, value string) bool {
	if _, ok := s.metrics.GaugeDict[name]; ok {
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return false
		}
		s.metrics.GaugeDict[name] = gauge(value)
		return true
	}
	return false
}

func (s MemoryStorage) GetGauge(name string) float64 {
	return float64(s.metrics.GaugeDict[name])
}

func (s MemoryStorage) StoreCounter(name string, value string) bool {
	if _, ok := s.metrics.CounterDict[name]; ok {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
		s.metrics.CounterDict[name] += counter(intVal)
		return true
	}
	return false
}

func (s MemoryStorage) GetCounter(name string) int64 {
	return int64(s.metrics.CounterDict[name])
}
