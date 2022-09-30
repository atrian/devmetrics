package storage

type (
	gauge   float64
	counter int64
)

type Repository interface {
	StoreGauge(name string, value string) bool
	GetGauge(name string) float64
	StoreCounter(name string, value string) bool
	GetCounter(name string) int64
}
