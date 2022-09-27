package storage

type MemoryStorage struct {
	metrics *MetricsDics
}

func NewMemoryStorage() *MemoryStorage {
	storage := MemoryStorage{metrics: NewMetricsDicts()}
	return &storage
}

func (s MemoryStorage) StoreGauge(name string, value gauge) bool {
	if _, ok := s.metrics.GaugeDict[name]; ok == true {
		s.metrics.GaugeDict[name] = value
		return true
	}
	return false
}

func (s MemoryStorage) GetGauge(name string) float64 {
	return float64(s.metrics.GaugeDict[name])
}

func (s MemoryStorage) StoreCounter(name string, value counter) bool {
	if _, ok := s.metrics.CounterDict[name]; ok == true {
		s.metrics.CounterDict[name] += value
		return true
	}
	return false
}

func (s MemoryStorage) GetCounter(name string) int64 {
	return int64(s.metrics.CounterDict[name])
}
