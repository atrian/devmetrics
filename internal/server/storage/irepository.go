package storage

type (
	gauge   float64
	counter int64
)

type repository interface {
	StoreGauge(name string, value gauge) bool
	GetGauge(name string) gauge
	StoreCounter(name string, value counter) bool
	GetCounter(name string) counter
}
