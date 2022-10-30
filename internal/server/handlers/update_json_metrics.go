package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/atrian/devmetrics/internal/dto"
)

// UpdateJSONMetrics обновление метрик POST /updates в JSON
func (h *Handler) UpdateJSONMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// список уникальных метрик в запросе
		countersRequested := make(map[string]int)
		gaugesRequested := make(map[string]int)

		metrics := unmarshallMetrics(r)
		verifiedMetrics := make([]dto.Metrics, 0, len(metrics))

		for _, metric := range metrics {
			switch metric.MType {
			case "gauge":
				// если валидация подписи нужна и она не прошла, пропускаем метрику
				if h.config.Server.HashKey != "" && !h.hasher.Compare(metric.Hash,
					fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value),
					h.config.Server.HashKey) {
					continue
				}
				gaugesRequested[metric.ID] += 1
			case "counter":
				if h.config.Server.HashKey != "" && !h.hasher.Compare(metric.Hash,
					fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta),
					h.config.Server.HashKey) {
					continue
				}
				countersRequested[metric.ID] += 1
			default:
				// непонятные метрики просто пропускаем
				continue
			}
			verifiedMetrics = append(verifiedMetrics, metric)
		}

		// сохраняем метрики с правильными подписями в БД
		h.storage.SetMetrics(verifiedMetrics)

		// слайс уникальных метрик для ответа с актуальными значениями
		responseMetrics := make([]dto.Metrics, 0, len(countersRequested)+len(gaugesRequested))

		// собираем актуальные значения counters
		for key := range countersRequested {
			actualCounterValue, _ := h.storage.GetCounter(key)

			metric := dto.Metrics{
				ID:    key,
				MType: "counter",
				Delta: &actualCounterValue,
			}

			// подписываем метрику если установлен ключ шифрования
			if h.config.Server.HashKey != "" {
				metric.Hash = h.hasher.Hash(fmt.Sprintf("%s:counter:%d", metric.ID, actualCounterValue),
					h.config.Server.HashKey)
			}

			responseMetrics = append(responseMetrics, metric)
		}

		// собираем актуальные значения gauges
		for key := range gaugesRequested {
			actualGaugeValue, _ := h.storage.GetGauge(key)

			metric := dto.Metrics{
				ID:    key,
				MType: "gauge",
				Value: &actualGaugeValue,
			}

			// подписываем метрику если установлен ключ шифрования
			if h.config.Server.HashKey != "" {
				metric.Hash = h.hasher.Hash(fmt.Sprintf("%s:gauge:%f", metric.ID, actualGaugeValue),
					h.config.Server.HashKey)
			}

			responseMetrics = append(responseMetrics, metric)
		}

		w.Header().Set("content-type", h.config.HTTP.ContentType)
		// устанавливаем статус-код 200
		w.WriteHeader(http.StatusOK)
		fmt.Println("Request OK")
		//json.NewEncoder(w).Encode(responseMetrics)
		testMetricVal := 0.6
		json.NewEncoder(w).Encode(dto.Metrics{
			ID:    "test",
			MType: "gauge",
			Value: &testMetricVal,
		})
	}
}

func unmarshallMetrics(r *http.Request) []dto.Metrics {
	var body io.Reader

	// если в заголовках установлен Content-Encoding gzip, распаковываем тело
	if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
		fmt.Println("here")
		body = decodeGzipBody(r.Body)
	} else {
		body = r.Body
	}

	var metrics []dto.Metrics
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&metrics)
	if err != nil {
		fmt.Println("JSON decode error:", err)
		fmt.Println("Metrics:", metrics)
	}

	return metrics
}

func decodeGzipBody(gzipR io.Reader) io.Reader {
	gz, err := gzip.NewReader(gzipR)
	if err != nil {
		fmt.Println("Error setting up gzip decoder", err)
	}
	return gz
}
