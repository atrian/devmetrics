package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/atrian/devmetrics/internal/dto"
)

// UpdateJSONMetric обновление метрик POST /update/ в JSON
// @Tags Metrics
// @Summary Обновление одной метрики с передачей данных в JSON формате
// @Accept  json
// @Produce json
// @Param metric body dto.Metrics true "Принимает JSON с данными метрики, возвращает JSON с обновленными данными"
// @Success 200 {object} dto.Metrics
// @Failure 400 {string} string ""
// @Failure 404 {string} string ""
// @Failure 500 {string} string ""
// @Router /update/ [post]
func (h *Handler) UpdateJSONMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := h.unmarshallMetric(r.Body)

		if err != nil {
			h.logger.Error("UpdateJSONMetric cant unmarshallMetric", err)
			http.Error(w, "Bad JSON", http.StatusBadRequest)
		}

		switch metric.MType {
		case "gauge":
			if h.config.Server.HashKey != "" && !h.hasher.Compare(metric.Hash,
				fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value),
				h.config.Server.HashKey) {
				http.Error(w, "Cant validate metric", http.StatusBadRequest)
			}

			h.storage.StoreGauge(metric.ID, *metric.Value)
			currentValue, _ := h.storage.GetGauge(metric.ID)
			metric.Value = &currentValue
		case "counter":
			if h.config.Server.HashKey != "" && !h.hasher.Compare(metric.Hash,
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

		h.logger.Debug(fmt.Sprintf("Request OK. metric %#v", metric))
		json.NewEncoder(w).Encode(metric)
	}
}

// unmarshallMetric анмаршаллинг метрики в dto.Metrics
func (h *Handler) unmarshallMetric(body io.ReadCloser) (*dto.Metrics, error) {
	var metric dto.Metrics
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&metric)

	if err != nil {
		return nil, err
	}

	return &metric, nil
}
