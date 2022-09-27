package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStorage_StoreCounter(t *testing.T) {
	storage := NewMemoryStorage()
	res := storage.StoreCounter("notAllowedKey", counter(1))

	// функция вернула false т.к. нет такого ключа в мапе
	assert.Equal(t, false, res)

	res = storage.StoreCounter("PollCount", counter(1))

	// функция вернула true т.к. ключ есть
	assert.Equal(t, true, res)
	// значение сохранилось в мапе
	assert.Equal(t, counter(1), storage.metrics.CounterDict["PollCount"])

	storage.StoreCounter("PollCount", counter(2022))
	// при последующем сохранении значение увеличилось а не перезаписалось 2022 + 1
	assert.Equal(t, counter(2023), storage.metrics.CounterDict["PollCount"])
}

func TestStorage_GetCounter(t *testing.T) {
	storage := NewMemoryStorage()
	storage.StoreCounter("PollCount", counter(1585))

	assert.Equal(t, int64(1585), storage.GetCounter("PollCount")) // значение можно получить через метод и оно совпадает с ожидаемым
}

func TestStorage_StoreGauge(t *testing.T) {
	storage := NewMemoryStorage()
	res := storage.StoreGauge("notAllowedKey", gauge(1))

	// функция вернула false т.к. нет такого ключа в мапе
	assert.Equal(t, false, res)

	res = storage.StoreGauge("Alloc", gauge(1))

	// функция вернула true т.к. ключ есть
	assert.Equal(t, true, res)
	// значение сохранилось в мапе по ключу Alloc
	assert.Equal(t, gauge(1), storage.metrics.GaugeDict["Alloc"])

	storage.StoreGauge("Alloc", gauge(777))
	// при последующем сохранении значение перезаписалось
	assert.Equal(t, gauge(777), storage.metrics.GaugeDict["Alloc"])
}

func TestStorage_GetGauge(t *testing.T) {
	storage := NewMemoryStorage()
	storage.StoreGauge("Alloc", gauge(777))

	assert.Equal(t, float64(777), storage.GetGauge("Alloc"))
}
