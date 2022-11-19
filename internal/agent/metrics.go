package agent

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/atrian/devmetrics/internal/dto"
)

type MetricsDics struct {
	GaugeDict   map[string]*GaugeMetric
	CounterDict map[string]*CounterMetric
	mu          sync.RWMutex
}

type GaugeMetric struct {
	value     gauge
	pullValue func(stats *runtime.MemStats) gauge
}

func (g *GaugeMetric) getGaugeValue() float64 {
	return float64(g.value)
}

type CounterMetric struct {
	value              counter
	calculateNextValue func(c *CounterMetric) counter
}

func (c *CounterMetric) getCounterValue() int64 {
	return int64(c.value)
}

func NewMetricsDicts() *MetricsDics {
	dict := MetricsDics{
		GaugeDict: map[string]*GaugeMetric{
			"Alloc": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.Alloc)
			}},
			"BuckHashSys": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.BuckHashSys)
			}},
			"Frees": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.Frees)
			}},
			"GCCPUFraction": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.GCCPUFraction)
			}},
			"GCSys": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.GCSys)
			}},
			"HeapAlloc": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapAlloc)
			}},
			"HeapIdle": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapIdle)
			}},
			"HeapInuse": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapInuse)
			}},
			"HeapObjects": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapObjects)
			}},
			"HeapReleased": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapReleased)
			}},
			"HeapSys": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.HeapSys)
			}},
			"LastGC": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.LastGC)
			}},
			"Lookups": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.Lookups)
			}},
			"MCacheInuse": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.MCacheInuse)
			}},
			"MCacheSys": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.MCacheSys)
			}},
			"MSpanInuse": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.MSpanInuse)
			}},
			"MSpanSys": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.MSpanSys)
			}},
			"Mallocs": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.Mallocs)
			}},
			"NextGC": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.NextGC)
			}},
			"NumForcedGC": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.NumForcedGC)
			}},
			"NumGC": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.NumGC)
			}},
			"OtherSys": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.OtherSys)
			}},
			"PauseTotalNs": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.PauseTotalNs)
			}},
			"StackInuse": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.StackInuse)
			}},
			"StackSys": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.StackSys)
			}},
			"Sys": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.Sys)
			}},
			"TotalAlloc": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(stats.TotalAlloc)
			}},
			"RandomValue": {pullValue: func(stats *runtime.MemStats) gauge {
				return gauge(rand.Float64())
			}},
		},
		CounterDict: map[string]*CounterMetric{
			"PollCount": {calculateNextValue: func(c *CounterMetric) counter {
				nextVal := c.value + 1
				return counter(nextVal)
			}},
		},
	}

	return &dict
}

func (md *MetricsDics) updateMetrics() {
	md.mu.Lock()         // блокируем mutex
	defer md.mu.Unlock() // разблокируем после обновления всех метрик

	// получаем данные мониторинга
	runtimeStats := runtime.MemStats{}
	runtime.ReadMemStats(&runtimeStats)

	// обновляем данные мониторинга по списку, обновляем счетчики
	for _, metric := range md.GaugeDict {
		metric.value = metric.pullValue(&runtimeStats)
	}
	for _, ct := range md.CounterDict {
		ct.value = ct.calculateNextValue(ct)
	}
}

// exportMetrics возвращает слайс DTO с подписанными метриками
func (md *MetricsDics) exportMetrics(sign func(metricType, id string, delta *int64, value *float64) string) *[]dto.Metrics {
	md.mu.RLock()         // берем mutex в режиме чтения
	defer md.mu.RUnlock() // разблокируем после выполнения

	exportedData := make([]dto.Metrics, 0, len(md.GaugeDict)+len(md.CounterDict))

	for key, metric := range md.GaugeDict {
		gaugeValue := metric.getGaugeValue()
		exportedData = append(exportedData, dto.Metrics{
			ID:    key,
			MType: "gauge",
			Delta: nil,
			Value: &gaugeValue,
			Hash:  sign("gauge", key, nil, &gaugeValue),
		})
	}

	for key, ct := range md.CounterDict {
		counterValue := ct.getCounterValue()
		exportedData = append(exportedData, dto.Metrics{
			ID:    key,
			MType: "counter",
			Delta: &counterValue,
			Value: nil,
			Hash:  sign("counter", key, &counterValue, nil),
		})
	}

	return &exportedData
}
