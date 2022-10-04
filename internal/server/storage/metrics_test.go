package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStorage_StoreCounter(t *testing.T) {
	storage := NewMemoryStorage()

	res := storage.StoreCounter("PollCount", "1")

	// функция вернула true т.к. ключ есть
	assert.Equal(t, true, res)
	// значение сохранилось в мапе
	assert.Equal(t, counter(1), storage.metrics.CounterDict["PollCount"])

	storage.StoreCounter("PollCount", "2022")
	// при последующем сохранении значение увеличилось а не перезаписалось 2022 + 1
	assert.Equal(t, counter(2023), storage.metrics.CounterDict["PollCount"])
}

func TestStorage_GetCounter(t *testing.T) {
	storage := NewMemoryStorage()
	storage.StoreCounter("PollCount", "1585")

	val, exist := storage.GetCounter("PollCount")
	assert.Equal(t, true, exist)
	assert.Equal(t, int64(1585), val) // значение можно получить через метод и оно совпадает с ожидаемым
}

func TestStorage_StoreGauge(t *testing.T) {
	storage := NewMemoryStorage()

	res := storage.StoreGauge("Alloc", "1")

	// функция вернула true т.к. ключ есть
	assert.Equal(t, true, res)
	// значение сохранилось в мапе по ключу Alloc
	assert.Equal(t, gauge(1), storage.metrics.GaugeDict["Alloc"])

	storage.StoreGauge("Alloc", "777")
	// при последующем сохранении значение перезаписалось
	assert.Equal(t, gauge(777), storage.metrics.GaugeDict["Alloc"])
}

func TestStorage_GetGauge(t *testing.T) {
	storage := NewMemoryStorage()
	storage.StoreGauge("Alloc", "777")

	val, exist := storage.GetGauge("Alloc")
	assert.Equal(t, true, exist)
	assert.Equal(t, float64(777), val)
}
