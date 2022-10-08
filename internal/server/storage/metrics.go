package storage

type MetricsDics struct {
	GaugeDict   map[string]gauge
	CounterDict map[string]counter
}

func NewMetricsDicts() *MetricsDics {
	dict := MetricsDics{
		GaugeDict:   map[string]gauge{},
		CounterDict: map[string]counter{},
	}

	return &dict
}
