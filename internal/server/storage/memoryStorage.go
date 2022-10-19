package storage

type MemoryStorage struct {
	metrics *MetricsDics
}

func NewMemoryStorage() *MemoryStorage {
	storage := MemoryStorage{metrics: NewMetricsDicts()}
	return &storage
}

func (s MemoryStorage) StoreGauge(name string, value float64) {
	s.metrics.GaugeDict[name] = gauge(value)
}

func (s MemoryStorage) GetGauge(name string) (float64, bool) {
	value, exist := s.metrics.GaugeDict[name]
	return float64(value), exist
}

func (s MemoryStorage) StoreCounter(name string, value int64) {
	s.metrics.CounterDict[name] += counter(value)
}

func (s MemoryStorage) GetCounter(name string) (int64, bool) {
	value, exist := s.metrics.CounterDict[name]
	return int64(value), exist
}

func (s MemoryStorage) GetMetrics() *MetricsDics {
	return s.metrics
}
