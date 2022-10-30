package storage

import "github.com/atrian/devmetrics/internal/dto"

type (
	gauge   float64
	counter int64
)

type Repository interface {
	StoreGauge(name string, value float64)
	GetGauge(name string) (float64, bool)
	StoreCounter(name string, value int64)
	GetCounter(name string) (int64, bool)
	GetMetrics() *MetricsDicts
	SetMetrics(metrics []dto.Metrics)
	Observer
}
