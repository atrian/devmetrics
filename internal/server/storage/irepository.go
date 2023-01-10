package storage

import "github.com/atrian/devmetrics/internal/dto"

type (
	gauge   float64
	counter int64
)

// IRepository интерфейс хранилища метрик на стороне сервера
type IRepository interface {
	StoreGauge(name string, value float64) error // StoreGauge запись метрики
	GetGauge(name string) (float64, bool)        // GetGauge получение значения метрики по имени
	StoreCounter(name string, value int64) error // StoreCounter запись счетчика
	GetCounter(name string) (int64, bool)        // GetCounter получение значения счетчика по имени
	GetMetrics() *MetricsDicts                   // GetMetrics получение всего справочника метрик MetricsDicts
	SetMetrics(metrics []dto.Metrics)            // SetMetrics массовое сохранение метрик из слайса dto.Metrics
	IObserver
}
