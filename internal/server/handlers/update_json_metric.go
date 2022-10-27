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
			if "" != h.config.Server.HashKey && false == h.hasher.Compare(metric.Hash,
				fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value),
				h.config.Server.HashKey) {
				http.Error(w, "Cant validate metric", http.StatusBadRequest)
			}

			h.storage.StoreGauge(metric.ID, *metric.Value)
			currentValue, _ := h.storage.GetGauge(metric.ID)
			metric.Value = &currentValue
		case "counter":
			if "" != h.config.Server.HashKey && false == h.hasher.Compare(metric.Hash,
				fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta),
				h.config.Server.HashKey) {
				http.Error(w, "Cant validate metric", http.StatusBadRequest)
			}

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
		fmt.Println(err) // TODO http.StatusBadRequest
	}

	return &metric
}
