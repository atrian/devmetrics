package storage

type Storage struct {
	metrics *MetricsDics
}

func NewStorage() *Storage {
	storage := Storage{metrics: NewMetricsDicts()}
	return &storage
}

func (s Storage) StoreGauge(name string, value gauge) bool {
	if _, ok := s.metrics.GaugeDict[name]; ok == true {
		s.metrics.GaugeDict[name] = value
		return true
	}
	return false
}

func (s Storage) GetGauge(name string) float64 {
	return float64(s.metrics.GaugeDict[name])
}

func (s Storage) StoreCounter(name string, value counter) bool {
	if _, ok := s.metrics.CounterDict[name]; ok == true {
		s.metrics.CounterDict[name] += value
		return true
	}
	return false
}

func (s Storage) GetCounter(name string) int64 {
	return int64(s.metrics.CounterDict[name])
}
