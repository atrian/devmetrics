// Package storage пакет содержит структуры для хранения метрик в памяти и постоянном хранилище
// интерфейс Repository и 2 его реализации MemoryStorage и PgSQLStorage
package storage

// MetricsDicts структура для хранения метрик и счетчиков
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
