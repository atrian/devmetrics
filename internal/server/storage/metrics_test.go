package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/atrian/devmetrics/internal/appconfig"
)

func TestStorage_StoreCounter(t *testing.T) {
	config := appconfig.NewServerConfig()
	storage := NewMemoryStorage(config)

	storage.StoreCounter("PollCount", int64(1))

	// значение сохранилось в мапе
	assert.Equal(t, counter(1), storage.metrics.CounterDict["PollCount"])

	storage.StoreCounter("PollCount", int64(2022))
	// при последующем сохранении значение увеличилось а не перезаписалось 2022 + 1
	assert.Equal(t, counter(2023), storage.metrics.CounterDict["PollCount"])
}

func TestStorage_GetCounter(t *testing.T) {
	config := appconfig.NewServerConfig()
	storage := NewMemoryStorage(config)
	storage.StoreCounter("PollCount", 1585)

	val, exist := storage.GetCounter("PollCount")
	assert.Equal(t, true, exist)
	assert.Equal(t, int64(1585), val) // значение можно получить через метод и оно совпадает с ожидаемым
}

func TestStorage_StoreGauge(t *testing.T) {
	config := appconfig.NewServerConfig()
	storage := NewMemoryStorage(config)

	storage.StoreGauge("Alloc", float64(1))

	// значение сохранилось в мапе по ключу Alloc
	assert.Equal(t, gauge(1), storage.metrics.GaugeDict["Alloc"])

	storage.StoreGauge("Alloc", float64(777))
	// при последующем сохранении значение перезаписалось
	assert.Equal(t, gauge(777), storage.metrics.GaugeDict["Alloc"])
}

func TestStorage_GetGauge(t *testing.T) {
	config := appconfig.NewServerConfig()
	storage := NewMemoryStorage(config)
	storage.StoreGauge("Alloc", float64(777))

	val, exist := storage.GetGauge("Alloc")
	assert.Equal(t, true, exist)
	assert.Equal(t, float64(777), val)
}
