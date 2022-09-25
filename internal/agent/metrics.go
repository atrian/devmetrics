package agent

import (
	"math/rand"
	"runtime"
)

type MetricsDics struct {
	GaugeDict   map[string]*GaugeMetric
	CounterDict map[string]*CounterMetric
}

type GaugeMetric struct {
	value     gauge
	initValue func(stats *runtime.MemStats) gauge
}

type CounterMetric struct {
	value     counter
	initValue func(c *CounterMetric) counter
}

func NewMetricsDicts() *MetricsDics {
	dict := MetricsDics{
		GaugeDict: map[string]*GaugeMetric{
			"Alloc": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.Alloc)
			}},
			"BuckHashSys": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.BuckHashSys)
			}},
			"Frees": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.Frees)
			}},
			"GCCPUFraction": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.GCCPUFraction)
			}},
			"GCSys": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.GCSys)
			}},
			"HeapAlloc": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapAlloc)
			}},
			"HeapIdle": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapIdle)
			}},
			"HeapInuse": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapInuse)
			}},
			"HeapObjects": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapObjects)
			}},
			"HeapReleased": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapReleased)
			}},
			"HeapSys": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapSys)
			}},
			"LastGC": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.LastGC)
			}},
			"Lookups": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.Lookups)
			}},
			"MCacheInuse": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.MCacheInuse)
			}},
			"MCacheSys": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.MCacheSys)
			}},
			"MSpanInuse": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.MSpanInuse)
			}},
			"MSpanSys": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.MSpanSys)
			}},
			"Mallocs": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.Mallocs)
			}},
			"NextGC": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.NextGC)
			}},
			"NumForcedGC": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.NumForcedGC)
			}},
			"NumGC": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.NumGC)
			}},
			"OtherSys": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.OtherSys)
			}},
			"PauseTotalNs": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.PauseTotalNs)
			}},
			"StackInuse": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.StackInuse)
			}},
			"StackSys": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.StackSys)
			}},
			"Sys": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.Sys)
			}},
			"TotalAlloc": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.TotalAlloc)
			}},
			"RandomValue": {initValue: func(stats *runtime.MemStats) gauge {
				return gauge(rand.Float64())
			}},
		},
		CounterDict: map[string]*CounterMetric{
			"PollCount": {initValue: func(c *CounterMetric) counter {
				nextVal := c.value + 1
				return counter(nextVal)
			}},
		},
	}

	return &dict
}
