package storage

type MetricsDicts struct {
	GaugeDict   map[string]gauge
	CounterDict map[string]counter
}

func NewMetricsDicts() *MetricsDicts {
	dict := MetricsDicts{
		GaugeDict:   map[string]gauge{},
		CounterDict: map[string]counter{},
	}

	return &dict
}
