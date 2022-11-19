package agent

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/atrian/devmetrics/internal/dto"
	"github.com/atrian/devmetrics/pkg/logger"
)

const (
	RuntimeMetric = iota
	GopsMetric
	CpuMetric
)

type MetricsDics struct {
	GaugeDict   map[string]*GaugeMetric
	CounterDict map[string]*CounterMetric
	logger      logger.Logger
	mu          sync.RWMutex
}

// StatsHolder контейнер для разных источников статистики
type StatsHolder struct {
	RuntimeMemStat *runtime.MemStats
	GopsMemStat    *mem.VirtualMemoryStat
	mu             sync.RWMutex
}

func NewStatsHolder() *StatsHolder {
	sh := StatsHolder{}
	sh.updateGopsMemStat()
	sh.updateRuntimeStat()

	return &sh
}

func (sh *StatsHolder) updateRuntimeStat() {
	sh.mu.Lock()
	var stat runtime.MemStats
	runtime.ReadMemStats(&stat)
	sh.RuntimeMemStat = &stat
	sh.mu.Unlock()
}

func (sh *StatsHolder) updateGopsMemStat() {
	sh.mu.Lock()
	sh.GopsMemStat, _ = mem.VirtualMemory()
	sh.mu.Unlock()
}

type GaugeMetric struct {
	source    int
	value     gauge
	pullValue func(sh *StatsHolder) gauge
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

func NewMetricsDicts(logger logger.Logger) *MetricsDics {
	dict := MetricsDics{
		GaugeDict: map[string]*GaugeMetric{
			"Alloc": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.Alloc)
			}},
			"BuckHashSys": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.BuckHashSys)
			}},
			"Frees": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.Frees)
			}},
			"GCCPUFraction": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.GCCPUFraction)
			}},
			"GCSys": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.GCSys)
			}},
			"HeapAlloc": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.HeapAlloc)
			}},
			"HeapIdle": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.HeapIdle)
			}},
			"HeapInuse": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.HeapInuse)
			}},
			"HeapObjects": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.HeapObjects)
			}},
			"HeapReleased": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.HeapReleased)
			}},
			"HeapSys": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.HeapSys)
			}},
			"LastGC": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.LastGC)
			}},
			"Lookups": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.Lookups)
			}},
			"MCacheInuse": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.MCacheInuse)
			}},
			"MCacheSys": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.MCacheSys)
			}},
			"MSpanInuse": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.MSpanInuse)
			}},
			"MSpanSys": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.MSpanSys)
			}},
			"Mallocs": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.Mallocs)
			}},
			"NextGC": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.NextGC)
			}},
			"NumForcedGC": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.NumForcedGC)
			}},
			"NumGC": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.NumGC)
			}},
			"OtherSys": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.OtherSys)
			}},
			"PauseTotalNs": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.PauseTotalNs)
			}},
			"StackInuse": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.StackInuse)
			}},
			"StackSys": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.StackSys)
			}},
			"Sys": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.Sys)
			}},
			"TotalAlloc": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.RuntimeMemStat.TotalAlloc)
			}},
			"RandomValue": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(rand.Float64())
			}},
			"TotalMemory": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.GopsMemStat.Total)
			}, source: GopsMetric},
			"FreeMemory": {pullValue: func(sh *StatsHolder) gauge {
				return gauge(sh.GopsMemStat.Free)
			}, source: GopsMetric},
		},
		CounterDict: map[string]*CounterMetric{
			"PollCount": {calculateNextValue: func(c *CounterMetric) counter {
				nextVal := c.value + 1
				return nextVal
			}},
		},
		logger: logger,
	}

	return &dict
}

func (md *MetricsDics) updateRuntimeMetrics() {
	md.mu.Lock()         // блокируем mutex
	defer md.mu.Unlock() // разблокируем после обновления всех метрик

	md.update(RuntimeMetric)
}

func (md *MetricsDics) updateGopsMetrics() {
	md.mu.Lock()         // блокируем mutex
	defer md.mu.Unlock() // разблокируем после обновления всех метрик

	md.updateCPUMetrics() // обновляем метрики CPU
	md.update(GopsMetric)
}

func (md *MetricsDics) updateCPUMetrics() {
	cpuStats, err := cpu.Percent(0, true)
	if err != nil {
		md.logger.Error("getCPUsStats cpu.Percent error", err)
	}

	for core, cpuPercent := range cpuStats {
		metricName := fmt.Sprintf("CPUutilization%v", core)
		md.GaugeDict[metricName] = &GaugeMetric{
			source: CpuMetric,
			value:  gauge(cpuPercent),
			pullValue: func(sh *StatsHolder) gauge {
				return gauge(0)
			},
		}
	}
}

func (md *MetricsDics) update(metricType int) {
	// получаем данные мониторинга
	statsHolder := NewStatsHolder()

	// обновляем данные мониторинга по списку с учетом источника
	for _, metric := range md.GaugeDict {
		switch metricType {
		case RuntimeMetric:
			if metric.source != RuntimeMetric {
				continue
			}
			metric.value = metric.pullValue(statsHolder)
		case GopsMetric:
			if metric.source != GopsMetric {
				continue
			}
			metric.value = metric.pullValue(statsHolder)
		}
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

	// выгружаем основные gauge метрики
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

	// выгружаем основные counter метрики
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
