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
			case "counter":
				if h.config.Server.HashKey != "" && !h.hasher.Compare(metric.Hash,
					fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta),
					h.config.Server.HashKey) {
					continue
				}
			default:
				// непонятные метрики просто пропускаем
				continue
			}
			verifiedMetrics = append(verifiedMetrics, metric)
		}

		// сохраняем метрики с правильными подписями в БД
		h.storage.SetMetrics(verifiedMetrics)

		w.Header().Set("content-type", h.config.HTTP.ContentType)
		// устанавливаем статус-код 200
		w.WriteHeader(http.StatusOK)
		fmt.Println("Request OK")
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
