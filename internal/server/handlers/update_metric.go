package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// UpdateMetric обновление метрик POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
// @Tags Metrics
// @Summary Запрос одной метрики с указанием её типа и имени
// @Produce html
// @Param metric_type path string true "Тип метрики: counter, gauge"
// @Param metric_name path string true "Имя метрики"
// @Param value path number true "Значение метрики"
// @Success 200 {number} number "Текущее значение метрики"
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /value/{metric_type}/{metric_name}/{value} [get]
func (h *Handler) UpdateMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		badRequestFlag := false
		var actualMetricValue string

		metricType := chi.URLParam(r, "metricType")
		metricTitle := chi.URLParam(r, "metricTitle")
		metricValue := chi.URLParam(r, "metricValue")

		if metricType != "gauge" && metricType != "counter" {
			http.Error(w, "Not implemented", http.StatusNotImplemented)
			return
		}

		h.logger.Debug(fmt.Sprintf("UpdateMetric chi.URLParam. metricType %v", metricType))

		if metricType == "gauge" {
			floatValue, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				badRequestFlag = true
			}

			err = h.storage.StoreGauge(metricTitle, floatValue)
			if err != nil {
				h.logger.Error("Cant store gauge metric", errors.Unwrap(err))
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			actualMetricValue = metricValue

			h.logger.Debug(fmt.Sprintf("Gauge metric stored. Metric: %v, actualMetricValue: %v", metricTitle, actualMetricValue))
		}

		if metricType == "counter" {
			intValue, err := strconv.Atoi(metricValue)
			if err != nil {
				badRequestFlag = true
			}

			err = h.storage.StoreCounter(metricTitle, int64(intValue))
			if err != nil {
				h.logger.Error("Cant store counter metric", errors.Unwrap(err))
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}

			counterVal, _ := h.storage.GetCounter(metricTitle)
			actualMetricValue = strconv.Itoa(int(counterVal))

			h.logger.Debug(fmt.Sprintf("Counter metric stored. Metric: %v, actualMetricValue: %v", metricTitle, actualMetricValue))
		}

		if badRequestFlag {
			h.logger.Info("Cant store metric, cant convert string value to storage value")
			http.Error(w, "Cant store metric", http.StatusBadRequest)
		} else {
			h.logger.Debug("Request OK")

			w.Header().Set("content-type", "text/plain")
			// устанавливаем статус-код 200
			w.WriteHeader(http.StatusOK)

			_, err := fmt.Fprint(w, actualMetricValue)
			if err != nil {
				h.logger.Error("UpdateMetric handler response writer", err)
			}
		}
	}
}
