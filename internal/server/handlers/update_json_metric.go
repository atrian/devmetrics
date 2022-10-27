package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/atrian/devmetrics/internal/dto"
)

// UpdateJSONMetric обновление метрик POST /update в JSON
func (h *Handler) UpdateJSONMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := unmarshallMetric(r.Body)

		switch metric.MType {
		case "gauge":
			h.storage.StoreGauge(metric.ID, *metric.Value)
			currentValue, _ := h.storage.GetGauge(metric.ID)
			metric.Value = &currentValue
		case "counter":
			h.storage.StoreCounter(metric.ID, *metric.Delta)
			currentValue, _ := h.storage.GetCounter(metric.ID)
			metric.Delta = &currentValue
		default:
			http.Error(w, "Cant store metric", http.StatusBadRequest)
		}

		w.Header().Set("content-type", h.config.HTTP.ContentType)
		// устанавливаем статус-код 200
		w.WriteHeader(http.StatusOK)

		fmt.Println("Request OK, current metric value:", metric)
		json.NewEncoder(w).Encode(metric)
	}
}

func unmarshallMetric(body io.ReadCloser) *dto.Metrics {
	var metric dto.Metrics
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&metric)
	if err != nil {
		panic(err)
	}

	return &metric
}
